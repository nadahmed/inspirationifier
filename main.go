package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type ImageDetails struct {
	Url  string
	Text string
}

func getImageFromUrl(w *http.ResponseWriter, url string, text string) (image.Image, string, error) {

	fontByte, err := os.ReadFile("Psilent.otf")

	if err != nil {
		http.Error(*w, "Image could not be downloaded", http.StatusNoContent)
		return nil, "", err
	}
	f, err := opentype.Parse(fontByte)

	if err != nil {
		http.Error(*w, "Font could not be loaded", http.StatusNoContent)
		return nil, "", err
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		http.Error(*w, "Font error", http.StatusFailedDependency)
		return nil, "", err
	}

	res, err := http.Get(url)

	if err != nil {
		http.Error(*w, "Image could not be downloaded", http.StatusNoContent)
		return nil, "", err
	}

	defer res.Body.Close()
	myImg, it, err := image.Decode(res.Body)

	if err != nil {
		fmt.Fprintln(*w, err.Error())
		return nil, "", err
	}

	// des := image.NewRGBA(myImg.Bounds())
	space := 8
	// var tbound fixed.Rectangle26_6
	// // var trect fixed.Int26_6
	// tbound, _ = font.BoundString(face, text)
	fmt.Println(face.Metrics().Height)
	des := image.NewRGBA(image.Rect(myImg.Bounds().Min.X, myImg.Bounds().Min.Y, myImg.Bounds().Dx(), myImg.Bounds().Dy()+24+(2*space)))
	draw.Draw(des, des.Bounds(), image.Black, myImg.Bounds().Min, draw.Src)
	draw.Draw(des, des.Bounds(), myImg, myImg.Bounds().Min, draw.Src)

	d := font.Drawer{
		Dst:  des,
		Src:  image.White,
		Face: face,
		Dot: fixed.P((des.Bounds().Dx()-font.MeasureString(face, text).Round())/2,
			des.Bounds().Dy()-space),
	}

	d.DrawString(text)

	return des, it, nil

}

func jsonHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		var id ImageDetails
		err := json.NewDecoder(r.Body).Decode(&id)

		if err != nil {
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		img, t, err := getImageFromUrl(&w, id.Url, id.Text)

		if err != nil {
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// buffer := new(bytes.Buffer)

		w.Header().Set("Content-Type", fmt.Sprintf("image/%s", t))
		// w.Header().Set("Content-Length", strconv.Itoa(int(unsafe.Sizeof(img))))
		// if _, err := w.Write(buffer.Bytes()); err != nil {
		// 	log.Println("unable to write image.")
		// }
		if err := png.Encode(w, img); err != nil {
			log.Println("unable to encode image.")
		}

		return
	}
	http.Error(w, r.Method+" Not allowed", http.StatusMethodNotAllowed)

}

func main() {
	http.HandleFunc("/get", jsonHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
