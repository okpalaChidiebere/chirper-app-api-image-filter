package imagefilterservice

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/okpalaChidiebere/chirper-app-api-image/data"
	imagefilterrepo "github.com/okpalaChidiebere/chirper-app-api-image/v0/image-filter/data_access"
)

// Custom type that allows setting the func that our Mock Do func will run instead
type MockDoType func(req *http.Request) (*http.Response, error)

// HTTPMockClient is the mock client
type HTTPMockClient struct {
 MockDo MockDoType
}
// Overriding what the Do function should "do" in our HTTPMockClient
func (m *HTTPMockClient) Do(req *http.Request) (*http.Response, error) {
 	return m.MockDo(req)
}

func Test_FilterImageFromURL(t *testing.T) {
	cases := []struct {
		name          string
		imageUrl string
		client func(t *testing.T) *HTTPMockClient
		buildStubs func(ctx context.Context, rrepoMock *imagefilterrepo.MockPresignerRepository)
		expectedServiceError error
	}{
		{
			name: "should return no error", 
			imageUrl: "https://itdoesntmatter.com/stupidName.jpg",
			client: func(t *testing.T) *HTTPMockClient {
				rect := image.Rect(0, 0, 200, 200)
				fakeImage := createRandomImage(rect)

				buf := new(bytes.Buffer)
				err := jpeg.Encode(buf, fakeImage, nil)
				if err != nil{
					t.Fatal(err)
				}

				send_s3 := buf.Bytes()
				// f, err := os.Open(data.Path("test/tyler.jpg"))
				// if err != nil{
				// 	t.Fatal(err)
				// }
				// defer f.Close()

				// img, _, err := image.Decode(f)
				// if err != nil{
				// 	t.Fatal(err)
				// }

				return &HTTPMockClient{
					MockDo: func(*http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body: io.NopCloser(bytes.NewReader(send_s3)),
						}, nil
					},
				}
			},
			buildStubs: func(ctx context.Context, repoMock *imagefilterrepo.MockPresignerRepository) {
				repoMock.EXPECT().GetGetSignedUrl(ctx, gomock.Any(), gomock.Any()).Times(0)
			},
		},
		{
			name: "should return no error and no repo error when image key is passed", 
			imageUrl: "stupidName.jpg",
			client: func(t *testing.T) *HTTPMockClient {
				rect := image.Rect(0, 0, 200, 200)
				fakeImage := createRandomImage(rect)

				buf := new(bytes.Buffer)
				err := jpeg.Encode(buf, fakeImage, nil)
				if err != nil{
					t.Fatal(err)
				}

				send_s3 := buf.Bytes()
				return &HTTPMockClient{
					MockDo: func(*http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body: io.NopCloser(bytes.NewReader(send_s3)),
						}, nil
					},
				}
			},
			buildStubs: func(ctx context.Context, repoMock *imagefilterrepo.MockPresignerRepository) {
				repoMock.EXPECT().GetGetSignedUrl(ctx, gomock.Any(), gomock.Any()).Times(1).Return("", nil)
			},
		},
		{
			name: "should return error when http request error occurs", 
			imageUrl: "https://itdoesntmatter.com/stupidImageKey.jpg",
			client: func(t *testing.T) *HTTPMockClient {

				return &HTTPMockClient {
					MockDo: func(*http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 404,
							Body: nil,
						}, errors.New("error fetching image")
					},
				}
			},
			buildStubs: func(ctx context.Context, repoMock *imagefilterrepo.MockPresignerRepository) {
				repoMock.EXPECT().GetGetSignedUrl(ctx, gomock.Any(), gomock.Any()).Times(0)
			},
			expectedServiceError: fmt.Errorf("cannot get from URL %v", errors.New("error fetching image")),
		},
	}

	var files []string

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.TODO()

			repoMock := imagefilterrepo.NewMockPresignerRepository(ctrl)
			tc.buildStubs(ctx, repoMock)

			service := New(repoMock)
			file, err := service.FilterImageFromURL(ctx, tc.imageUrl, tc.client(t))

			_, ok := err.(*fs.PathError)

			// in Travis CI, you always get an error when trying go is trying to create a file. So we can ignore path errors
			if !ok {
				assert.Equal(t, tc.expectedServiceError, err)
			}

			if(file != ""){
				files = append(files, file)
			}
		})
	}

	//clean up test files
	DeleteLocalFiles(files)
}

func Test_UploadImageToS3Bucket(t *testing.T) {
	//create a tmp file
	file, err := os.CreateTemp(data.Path("test"), "tyler*.jpg") // For example "test/tyler54003078.jpg"
    if err != nil {
        fmt.Println(err)
    }
	defer file.Close()

	//create a fake image
	rect := image.Rect(0, 0, 200, 200)
	fakeImage := createRandomImage(rect)

	buf := new(bytes.Buffer)
	_ = jpeg.Encode(buf, fakeImage, nil)
	send_s3 := buf.Bytes()

	//write the new random fake image to that file to be used to actually test the UploadImageToS3Bucket method
    if _, err := file.Write(send_s3); err != nil {
        fmt.Println(err)
    }

	fakeImageFilepath := file.Name()
	cases := []struct {
		name          string
		client func() *HTTPMockClient
		filePath string
		expectObjectKey string
		expectedServiceError error
	}{
		{
			name: "Should return no error",
			client: func() *HTTPMockClient {
				return &HTTPMockClient{
					MockDo: func(*http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body: io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil
					},
				}
			},
			filePath: fakeImageFilepath,
			expectObjectKey: filepath.Base(fakeImageFilepath),
		},
		{
			name: "Should return error as http request fails",
			client: func() *HTTPMockClient {
				return &HTTPMockClient{
					MockDo: func(*http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 500,
							Body:       nil,
						}, errors.New("Error from aws web server")
					},
				}
			},
			filePath: fakeImageFilepath,
			expectedServiceError: errors.New("Error from aws web server"),
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.TODO()

			repoMock := imagefilterrepo.NewMockPresignerRepository(ctrl)
			tokens := strings.Split(tc.filePath, "/")
			repoMock.EXPECT().GetPutSignedUrl(ctx, gomock.Eq("chirper-app-thumbnail-dev"), gomock.Eq(tokens[len(tokens) - 1])).Return("", nil).AnyTimes() //come back to this and set this i time

			service := New(repoMock)
			key, err := service.UploadImageToS3Bucket(ctx, tc.filePath, tc.client())

			assert.EqualValues(t, tc.expectObjectKey, key)
			assert.Equal(t, tc.expectedServiceError, err)
		})
	}

	// We can choose to have these files deleted on program close
    // defer os.Remove(file.Name())
	DeleteLocalFiles([]string{file.Name()})
}

func createRandomImage(rect image.Rectangle) (created *image.NRGBA) {
	pix := make([]uint8, rect.Dx()*rect.Dy()*4)
	rand.Read(pix)
	created = &image.NRGBA{
		Pix:    pix,
		Stride: rect.Dx() * 4,
		Rect:   rect,
	}
	return
}
