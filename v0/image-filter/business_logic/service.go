package imagefilterservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/okpalaChidiebere/chirper-app-api-image/config"
	"github.com/okpalaChidiebere/chirper-app-api-image/data"
	"github.com/okpalaChidiebere/chirper-app-api-image/v0/common"
	repo "github.com/okpalaChidiebere/chirper-app-api-image/v0/image-filter/data_access"
)

var (
	mConfig    = config.NewConfig()
)

// Learn  more on custom reader here
// https://medium.com/learning-the-go-programming-language/streaming-io-in-go-d93507931185
type ProgressReader struct {
    io.Reader //could be of type os.File, bytes.Buffer or io.ReadCloser because they both satisfies the io.Reader interface
    OnUploadProgress func(r int64)
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
    n, err = pr.Reader.Read(p)
    pr.OnUploadProgress(int64(n))
    return
}

type ServiceImpl struct {
	repo repo.PresignerRepository
}

func New(repo repo.PresignerRepository) *ServiceImpl {
	return &ServiceImpl{repo}
}

func (s *ServiceImpl) GetGetSignedUrl(ctx context.Context, bucketName, objectKey string) (url string, err error) {
    return s.repo.GetGetSignedUrl(ctx,bucketName, objectKey)
}

func (s *ServiceImpl) FilterImageFromURL(ctx context.Context, inputURL string, httpRequester common.IHttpRequester) (filteredUrl string, err error){
    imageKey := uuid.NewString()
    _, err = url.ParseRequestURI(inputURL) //check see if the url is a singed if a publicly accessible url or an imageKey
    if err != nil {
        //we can assume its an aws image key
        url, err := s.repo.GetGetSignedUrl(ctx, mConfig.Dev.ImageBucket, inputURL)
        if err != nil {
            return "", err
        }
        imageKey = inputURL
        inputURL = url
    }

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, inputURL, nil)
	r, err := httpRequester.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot get from URL %v", err)
	}
	if r != nil {
		defer r.Body.Close()
	}

	if r.StatusCode != 200 {
        return "", errors.New("received non 200 response code")
    }
	
	img, _, err := image.Decode(r.Body)
    if err != nil {
        return "", err
    }

	// outPath := data.Path(fmt.Sprintf("tmp/filtered.%s.jpg", imageKey))

    //this is the resized image
    resImg := resize(img, 300, 300)

	//grey scale the image
	result := image.NewGray(resImg.Bounds())
	draw.Draw(result, result.Bounds(), resImg, resImg.Bounds().Min, draw.Src)

    //this is the resized image []bytes with desired quality
    imgBytes := quality(result, 90)

    // f, err := os.Create(outPath)
    f, err :=  os.CreateTemp(data.Path("tmp"), fmt.Sprintf("filtered.%s.*jpg", imageKey))
    if err != nil {
        return "", err
    }

    _, err = f.Write(imgBytes)
    if err != nil {
        return "", err
    }

	// err = os.WriteFile(outPath, imgBytes, 0777)
    // if err != nil {
    //     return "", err
    // }

    outPath := f.Name()

	return outPath, nil
}

func (s *ServiceImpl) UploadImageToS3Bucket(ctx context.Context, filePath string, httpRequester common.IHttpRequester) (imageKey string, err error){
	key := filepath.Base(filePath)

	file, err := os.Open(data.Path(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to open file %v, %v", key, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

    loaded := int64(0)
    pr := &ProgressReader{file, func(r int64) {
        loaded += r
        if r > 0 {
            log.Printf("\rtotal read:%d; progress:%d%%", loaded, int(float32(loaded*100)/float32(fileInfo.Size())))
        } else{
            //done
            log.Printf("sent %d of %d bytes for data %s \n", loaded, fileInfo.Size(), fileInfo.Name())
        }
    }}
    
	url, err := s.repo.GetPutSignedUrl(ctx, mConfig.Dev.ImageBucket, key)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, pr )
    req.ContentLength = fileInfo.Size()

	if err != nil {
		return "", err
	}
	res, err := httpRequester.Do(req)
	if err != nil {
		return "", err
	}
	// Close response body as required.
    if res != nil {
		defer res.Body.Close()
	}

	log.Printf("%v", res.StatusCode)
	
	return key, nil
}

func DeleteLocalFiles(files []string){
    for _, file := range files {
        if err := os.Remove(file); err != nil {
            log.Printf("cannot delete file: %v", err)
        }
    }
}

func resize(img image.Image, length int, width int) image.Image {
    //truncate pixel size
    minX := img.Bounds().Min.X
    minY := img.Bounds().Min.Y
    maxX := img.Bounds().Max.X
    maxY := img.Bounds().Max.Y
    for (maxX-minX)%length != 0 {
        maxX--
    }
    for (maxY-minY)%width!= 0 {
        maxY--
    }
    scaleX := (maxX - minX) / length
    scaleY := (maxY - minY) / width

    imgRect := image.Rect(0, 0, length, width)
    resImg := image.NewRGBA(imgRect)
    draw.Draw(resImg, resImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
    for y := 0; y < width; y += 1 {
        for x := 0; x < length; x += 1 {
            averageColor := getAverageColor(img, minX+x*scaleX, minX+(x+1)*scaleX, minY+y*scaleY, minY+(y+1)*scaleY)
            resImg.Set(x, y, averageColor)
        }
    }
    return resImg
}

func getAverageColor(img image.Image, minX int, maxX int, minY int, maxY int) color.Color {
    var averageRed float64
    var averageGreen float64
    var averageBlue float64
    var averageAlpha float64
    scale := 1.0 / float64((maxX-minX)*(maxY-minY))

    for i := minX; i < maxX; i++ {
        for k := minY; k < maxY; k++ {
            r, g, b, a := img.At(i, k).RGBA()
            averageRed += float64(r) * scale
            averageGreen += float64(g) * scale
            averageBlue += float64(b) * scale
            averageAlpha += float64(a) * scale
        }
    }

    averageRed = math.Sqrt(averageRed)
    averageGreen = math.Sqrt(averageGreen)
    averageBlue = math.Sqrt(averageBlue)
    averageAlpha = math.Sqrt(averageAlpha)

    averageColor := color.RGBA{
        R: uint8(averageRed),
        G: uint8(averageGreen),
        B: uint8(averageBlue),
        A: uint8(averageAlpha)}

    return averageColor
}

func quality(img image.Image, quality int) []byte {
    var opt jpeg.Options
    opt.Quality = quality //image quality

    buff := bytes.NewBuffer(nil)
    err := jpeg.Encode(buff, img, &opt)
    if err != nil {
        log.Fatal(err)
    }

    return buff.Bytes()
}

// func buildFileName(fullUrlFile string) (string, error) {
//     fileUrl, e := url.Parse(fullUrlFile)
//     if e != nil {
//        return "", e
//     }

//     path := fileUrl.Path
//     segments := strings.Split(path, "/")

//     fileName := segments[len(segments)-1]
//     // println(fileName)
// 	return fileName, nil
// }