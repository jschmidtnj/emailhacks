package main

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func validateContentType(thetype string) error {
	for _, validtype := range validContentTypes {
		if validtype == thetype {
			return nil
		}
	}
	return errors.New("invalid content type provided")
}

func uploadFile(fileBuffer *bytes.Buffer, filewriter *storage.Writer) (int64, error) {
	byteswritten, err := fileBuffer.WriteTo(filewriter)
	if err != nil {
		return -1, errors.New("error writing to filewriter: " + err.Error())
	}
	err = filewriter.Close()
	if err != nil {
		return -1, errors.New("error closing writer: " + err.Error())
	}
	return byteswritten, nil
}

func writeGenericFile(file io.Reader, filetype string, posttype string, fileidDecoded string, postid string) (int64, error) {
	var filebuffer bytes.Buffer
	io.Copy(&filebuffer, file)
	defer filebuffer.Reset()
	var fileobj *storage.ObjectHandle
	var fileIndex string
	if posttype == formType {
		fileIndex = formFileIndex
	} else if posttype == responseType {
		fileIndex = responseFileIndex
	} else {
		fileIndex = blogFileIndex
	}
	fileobj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + originalPath)
	filewriter := fileobj.NewWriter(ctxStorage)
	filewriter.ContentType = filetype
	filewriter.Metadata = map[string]string{}
	byteswritten, err := uploadFile(&filebuffer, filewriter)
	if err != nil {
		return -1, err
	}
	return byteswritten, nil
}

func writeJpeg(file io.Reader, filetype string, posttype string, fileidDecoded string, postid string) (int64, error) {
	originalImage, _, err := image.Decode(file)
	if err != nil {
		return -1, err
	}
	originalImageBuffer := new(bytes.Buffer)
	defer originalImageBuffer.Reset()
	jpegOptionsOriginal := jpeg.Options{Quality: 90}
	err = jpeg.Encode(originalImageBuffer, originalImage, &jpegOptionsOriginal)
	if err != nil {
		return -1, err
	}
	blurredImage := imaging.Blur(originalImage, progressiveImageBlurAmount)
	blurredImageBuffer := new(bytes.Buffer)
	defer blurredImageBuffer.Reset()
	jpegOptionsBlurred := jpeg.Options{Quality: 60}
	err = jpeg.Encode(blurredImageBuffer, blurredImage, &jpegOptionsBlurred)
	if err != nil {
		return -1, err
	}
	var originalImageObj *storage.ObjectHandle
	var blurredImageObj *storage.ObjectHandle
	var fileIndex string
	if posttype == formType {
		fileIndex = formFileIndex
	} else if posttype == responseType {
		fileIndex = responseFileIndex
	} else {
		fileIndex = blogFileIndex
	}
	originalImageObj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + originalPath)
	blurredImageObj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + blurPath)
	originalImageWriter := originalImageObj.NewWriter(ctxStorage)
	originalImageWriter.ContentType = filetype
	originalImageWriter.Metadata = map[string]string{}
	byteswritten, err := uploadFile(originalImageBuffer, originalImageWriter)
	if err != nil {
		return -1, err
	}
	blurredImageWriter := blurredImageObj.NewWriter(ctxStorage)
	blurredImageWriter.ContentType = filetype
	blurredImageWriter.Metadata = map[string]string{}
	morebyteswritten, err := uploadFile(blurredImageBuffer, blurredImageWriter)
	if err != nil {
		return -1, err
	}
	byteswritten += morebyteswritten
	return byteswritten, nil
}

func writePng(file io.Reader, filetype string, posttype string, fileidDecoded string, postid string) (int64, error) {
	originalImage, _, err := image.Decode(file)
	if err != nil {
		return -1, err
	}
	originalImageBuffer := new(bytes.Buffer)
	defer originalImageBuffer.Reset()
	err = png.Encode(originalImageBuffer, originalImage)
	if err != nil {
		return -1, err
	}
	blurredImage := imaging.Blur(originalImage, progressiveImageBlurAmount)
	blurredImageBuffer := new(bytes.Buffer)
	defer blurredImageBuffer.Reset()
	err = png.Encode(blurredImageBuffer, blurredImage)
	if err != nil {
		return -1, err
	}
	var originalImageObj *storage.ObjectHandle
	var blurredImageObj *storage.ObjectHandle
	var fileIndex string
	if posttype == formType {
		fileIndex = formFileIndex
	} else if posttype == responseType {
		fileIndex = responseFileIndex
	} else {
		fileIndex = blogFileIndex
	}
	originalImageObj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + originalPath)
	blurredImageObj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + blurPath)
	originalImageWriter := originalImageObj.NewWriter(ctxStorage)
	originalImageWriter.ContentType = filetype
	originalImageWriter.Metadata = map[string]string{}
	byteswritten, err := uploadFile(originalImageBuffer, originalImageWriter)
	if err != nil {
		return -1, err
	}
	blurredImageWriter := blurredImageObj.NewWriter(ctxStorage)
	blurredImageWriter.ContentType = filetype
	blurredImageWriter.Metadata = map[string]string{}
	morebyteswritten, err := uploadFile(blurredImageBuffer, blurredImageWriter)
	if err != nil {
		return -1, err
	}
	byteswritten += morebyteswritten
	return byteswritten, nil
}

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int
	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}
	return highestX - lowestX, highestY - lowestY
}

func writeGif(file io.Reader, filetype string, posttype string, fileidDecoded string, postid string) (int64, error) {
	originalGif, err := gif.DecodeAll(file)
	if err != nil {
		return -1, err
	}
	originalGifBuffer := new(bytes.Buffer)
	defer originalGifBuffer.Reset()
	err = gif.EncodeAll(originalGifBuffer, originalGif)
	if err != nil {
		return -1, err
	}
	imgWidth, imgHeight := getGifDimensions(originalGif)
	originalImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(originalImage, originalImage.Bounds(), originalGif.Image[0], image.ZP, draw.Src)
	originalImageBuffer := new(bytes.Buffer)
	defer originalImageBuffer.Reset()
	jpegOptionsOriginal := jpeg.Options{Quality: 90}
	err = jpeg.Encode(originalImageBuffer, originalImage, &jpegOptionsOriginal)
	if err != nil {
		return -1, err
	}
	blurredImage := imaging.Blur(originalImage, progressiveImageBlurAmount)
	blurredImageBuffer := new(bytes.Buffer)
	defer blurredImageBuffer.Reset()
	jpegOptionsBlurred := jpeg.Options{Quality: 60}
	err = jpeg.Encode(blurredImageBuffer, blurredImage, &jpegOptionsBlurred)
	if err != nil {
		return -1, err
	}
	var originalGifObj *storage.ObjectHandle
	var originalImageObj *storage.ObjectHandle
	var blurredImageObj *storage.ObjectHandle
	var fileIndex string
	if posttype == formType {
		fileIndex = formFileIndex
	} else if posttype == responseType {
		fileIndex = responseFileIndex
	} else {
		fileIndex = blogFileIndex
	}
	originalGifObj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + originalPath)
	originalImageObj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + placeholderPath + originalPath)
	blurredImageObj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileidDecoded + placeholderPath + blurPath)
	originalGifWriter := originalGifObj.NewWriter(ctxStorage)
	originalGifWriter.ContentType = filetype
	originalGifWriter.Metadata = map[string]string{}
	byteswritten, err := uploadFile(originalGifBuffer, originalGifWriter)
	if err != nil {
		return -1, err
	}
	var placeholderFileType = "image/jpeg"
	originalImageWriter := originalImageObj.NewWriter(ctxStorage)
	originalImageWriter.ContentType = placeholderFileType
	originalImageWriter.Metadata = map[string]string{}
	morebyteswritten, err := uploadFile(originalImageBuffer, originalImageWriter)
	if err != nil {
		return -1, err
	}
	byteswritten += morebyteswritten
	blurredImageWriter := blurredImageObj.NewWriter(ctxStorage)
	blurredImageWriter.ContentType = placeholderFileType
	blurredImageWriter.Metadata = map[string]string{}
	morebyteswritten, err = uploadFile(blurredImageBuffer, blurredImageWriter)
	if err != nil {
		return -1, err
	}
	byteswritten += morebyteswritten
	return byteswritten, nil
}

func writeFile(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != http.MethodPut {
		handleError("upload file http method not PUT", http.StatusBadRequest, response)
		return
	}
	filetype := request.URL.Query().Get("filetype")
	if filetype == "" {
		handleError("error getting filetype from query", http.StatusBadRequest, response)
		return
	}
	if err := validateContentType(filetype); err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	posttype := request.URL.Query().Get("posttype")
	if posttype == "" {
		handleError("error getting post type from query", http.StatusBadRequest, response)
		return
	}
	if !findInArray(posttype, validStorageTypes) {
		handleError("invalid posttype in query", http.StatusBadRequest, response)
		return
	}
	postid := request.URL.Query().Get("postid")
	if postid == "" {
		handleError("error getting post id from query", http.StatusBadRequest, response)
		return
	}
	postIDObj, err := primitive.ObjectIDFromHex(postid)
	if err != nil {
		handleError("invalid post id found", http.StatusBadRequest, response)
		return
	}
	ownerID, err := validateStorageEditRequest(request, postIDObj, posttype)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	var fileid string
	if posttype == blogType {
		fileid = request.URL.Query().Get("fileid")
		if fileid == "" {
			handleError("error getting file id from query", http.StatusBadRequest, response)
			return
		}
	} else {
		uuid, err := uuid.NewRandom()
		if err != nil {
			handleError("error creating uuid: "+err.Error(), http.StatusBadRequest, response)
			return
		}
		fileid = uuid.String()
	}
	file, fileHeader, err := request.FormFile("file")
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	defer file.Close()
	if posttype == responseType || posttype == formType {
		userData, err := getAccount(ownerID, false)
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		usedStorage := userData["storage"].(int)
		productData, err := getProductFromUserData(userData)
		if err != nil {
			handleError(err.Error(), http.StatusUnauthorized, response)
			return
		}
		maxStorage := productData["maxstorage"].(int)
		storageRemaining := maxStorage - usedStorage
		if fileHeader.Size > int64(storageRemaining) {
			handleError("not enough storage remaining", http.StatusBadRequest, response)
			return
		}
	}
	var byteswritten int64
	switch filetype {
	case "image/jpeg":
		if byteswritten, err = writeJpeg(file, filetype, posttype, fileid, postid); err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		break
	case "image/png":
		if byteswritten, err = writePng(file, filetype, posttype, fileid, postid); err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		break
	case "image/gif":
		if byteswritten, err = writeGif(file, filetype, posttype, fileid, postid); err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		break
	default:
		if byteswritten, err = writeGenericFile(file, filetype, posttype, fileid, postid); err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		break
	}
	if posttype == responseType || posttype == formType {
		_, err = userCollection.UpdateOne(ctxMongo, bson.M{
			"_id": ownerID,
		}, bson.M{
			"$inc": bson.M{
				"storage": byteswritten,
			},
		})
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
	}
	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(`{"message":"file updated","id":"` + fileid + `"}`))
}

func deleteFiles(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != http.MethodDelete {
		handleError("delete post files http method not Delete", http.StatusBadRequest, response)
		return
	}
	authToken := getAuthToken(request)
	if _, err := getTokenData(authToken); err != nil {
		handleError("auth error: "+err.Error(), http.StatusBadRequest, response)
		return
	}
	var filedata map[string]interface{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		handleError("error getting request body: "+err.Error(), http.StatusBadRequest, response)
		return
	}
	err = json.Unmarshal(body, &filedata)
	if err != nil {
		handleError("error parsing request body: "+err.Error(), http.StatusBadRequest, response)
		return
	}
	if filedata["fileids"] == nil {
		handleError("no fileids provided", http.StatusBadRequest, response)
		return
	}
	if filedata["postid"] == nil {
		handleError("no postid provided", http.StatusBadRequest, response)
		return
	}
	if filedata["posttype"] == nil {
		handleError("no post type provided", http.StatusBadRequest, response)
		return
	}
	postid, ok := filedata["postid"].(string)
	if !ok {
		handleError("unable to cast post id to string", http.StatusBadRequest, response)
		return
	}
	posttype, ok := filedata["posttype"].(string)
	if !ok {
		handleError("unable to cast posttype to string", http.StatusBadRequest, response)
		return
	}
	if !findInArray(posttype, validStorageTypes) {
		handleError("invalid posttype in body", http.StatusBadRequest, response)
		return
	}
	postIDObj, err := primitive.ObjectIDFromHex(postid)
	if err != nil {
		handleError("invalid post id found", http.StatusBadRequest, response)
		return
	}
	if posttype == formType {
		_, err = checkFormAccess(postIDObj, authToken, editAccessLevel, false)
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
	}
	fileids, ok := filedata["fileids"].([]interface{})
	if !ok {
		handleError("file ids cannot be cast to interface array", http.StatusBadRequest, response)
		return
	}
	ownerID, err := validateStorageEditRequest(request, postIDObj, posttype)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	var bytesRemoved int64 = 0
	for _, fileidinterface := range fileids {
		fileid, ok := fileidinterface.(string)
		if !ok {
			handleError("file id cannot be cast to string", http.StatusBadRequest, response)
			return
		}
		var fileobj *storage.ObjectHandle
		var fileIndex string
		if posttype == formType {
			fileIndex = formFileIndex
		} else if posttype == responseType {
			fileIndex = responseFileIndex
		} else {
			fileIndex = blogFileIndex
		}
		fileobj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileid + originalPath)
		fileobjattributes, err := fileobj.Attrs(ctxStorage)
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		var filetype = fileobjattributes.ContentType
		var hasblur = false
		for _, blurtype := range haveblur {
			if blurtype == filetype {
				hasblur = true
				break
			}
		}
		bytesRemoved += fileobjattributes.Size
		if err := fileobj.Delete(ctxStorage); err != nil {
			handleError("error deleting original file: "+err.Error(), http.StatusBadRequest, response)
			return
		}
		if hasblur {
			var fileIndex string
			if posttype == formType {
				fileIndex = formFileIndex
			} else if posttype == responseType {
				fileIndex = responseFileIndex
			} else {
				fileIndex = blogFileIndex
			}
			var addPath = ""
			if filetype == "image/gif" {
				fileobj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileid + placeholderPath + originalPath)
				fileobjattributes, err := fileobj.Attrs(ctxStorage)
				if err != nil {
					handleError(err.Error(), http.StatusBadRequest, response)
					return
				}
				bytesRemoved += fileobjattributes.Size
				if err := fileobj.Delete(ctxStorage); err != nil {
					handleError("error deleting blur file: "+err.Error(), http.StatusBadRequest, response)
					return
				}
				addPath = placeholderPath
			}
			fileobj = storageBucket.Object(fileIndex + "/" + postid + "/" + fileid + addPath + blurPath)
			fileobjattributes, err := fileobj.Attrs(ctxStorage)
			if err != nil {
				handleError(err.Error(), http.StatusBadRequest, response)
				return
			}
			bytesRemoved += fileobjattributes.Size
			if err := fileobj.Delete(ctxStorage); err != nil {
				handleError("error deleting blur file: "+err.Error(), http.StatusBadRequest, response)
				return
			}
		}
	}
	if posttype == responseType || posttype == formType {
		_, err = userCollection.UpdateOne(ctxMongo, bson.M{
			"_id": ownerID,
		}, bson.M{
			"$dec": bson.M{
				"storage": bytesRemoved,
			},
		})
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
	}
	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(`{"message":"files deleted"}`))
}

func validateStorageEditRequest(request *http.Request, postID primitive.ObjectID, posttype string) (primitive.ObjectID, error) {
	accessToken := request.URL.Query().Get("updateToken")
	postIDString := postID.Hex()
	var ownerString string
	var err error
	if len(accessToken) > 0 && posttype == responseType {
		// validate access token
		var tokenResponseIDString string
		tokenResponseIDString, ownerString, _, err = getResponseEditTokenData(accessToken, editAccessLevel)
		if err != nil {
			return primitive.NilObjectID, err
		}
		_, err = primitive.ObjectIDFromHex(tokenResponseIDString)
		if err != nil {
			return primitive.NilObjectID, err
		}
		if tokenResponseIDString != postIDString {
			return primitive.NilObjectID, err
		}
	} else {
		accessToken = getAuthToken(request)
		if _, err := getTokenData(accessToken); err != nil {
			return primitive.NilObjectID, err
		}
		_, err = validateAdmin(accessToken)
		isAdmin := err == nil
		if !isAdmin {
			if posttype == blogType {
				return primitive.NilObjectID, errors.New("you need to be admin to edit blogs")
			} else if posttype == formType {
				formData, err := checkFormAccess(postID, accessToken, editAccessLevel, false)
				if err != nil {
					return primitive.NilObjectID, err
				}
				ownerString = formData.Owner
			} else if posttype == responseType {
				responseData, err := checkResponseAccess(postID, accessToken, editAccessLevel, false)
				if err != nil {
					return primitive.NilObjectID, err
				}
				ownerString = responseData["owner"].(string)
			}
		} else if posttype == responseType {
			responseData, err := getResponse(postID, false)
			if err != nil {
				return primitive.NilObjectID, err
			}
			ownerString = responseData["owner"].(string)
		} else if posttype == formType {
			formData, err := getForm(postID, false)
			if err != nil {
				return primitive.NilObjectID, err
			}
			ownerString = formData.Owner
		}
	}
	ownerID, err := primitive.ObjectIDFromHex(ownerString)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return ownerID, nil
}

func getFile(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != http.MethodGet {
		handleError("get post file http method not GET", http.StatusBadRequest, response)
		return
	}
	posttype := request.URL.Query().Get("posttype")
	if posttype == "" {
		handleError("error getting posttype from query", http.StatusBadRequest, response)
		return
	}
	if !findInArray(posttype, validStorageTypes) {
		handleError("invalid posttype in query", http.StatusBadRequest, response)
		return
	}
	postid := request.URL.Query().Get("postid")
	if postid == "" {
		handleError("error getting post id from query", http.StatusBadRequest, response)
		return
	}
	postIDObj, err := primitive.ObjectIDFromHex(postid)
	if err != nil {
		handleError("invalid post id: "+err.Error(), http.StatusBadRequest, response)
		return
	}
	fileid := request.URL.Query().Get("fileid")
	if fileid == "" {
		handleError("no file id", http.StatusBadRequest, response)
		return
	}
	requestType := request.URL.Query().Get("requestType")
	if requestType == "" {
		handleError("error getting request type from query", http.StatusBadRequest, response)
		return
	}
	if requestType != "original" && requestType != "placeholder" && requestType != "blur" {
		handleError("invalid request type given", http.StatusBadRequest, response)
		return
	}
	fileType := request.URL.Query().Get("fileType")
	if fileType == "" {
		handleError("error getting file type from query", http.StatusBadRequest, response)
		return
	}
	updateToken := request.URL.Query().Get("updateToken")
	if posttype == formType {
		if updateToken == "" {
			accessToken := getAuthToken(request)
			_, err := checkFormAccess(postIDObj, accessToken, viewAccessLevel, false)
			if err != nil {
				handleError("no form access: "+err.Error(), http.StatusUnauthorized, response)
				return
			}
		} else {
			tokenFormIDString, _, _, err := getFormUpdateClaimsData(updateToken, viewAccessLevel)
			if err != nil {
				handleError("invalid view token: "+err.Error(), http.StatusUnauthorized, response)
				return
			}
			if tokenFormIDString != postid {
				handleError("view token for wrong form: "+err.Error(), http.StatusUnauthorized, response)
				return
			}
		}
	} else if posttype == responseType {
		if updateToken == "" {
			accessToken := getAuthToken(request)
			_, err := checkResponseAccess(postIDObj, accessToken, viewAccessLevel, false)
			if err != nil {
				handleError("no response access: "+err.Error(), http.StatusUnauthorized, response)
				return
			}
		} else {
			tokenResponseIDString, _, _, err := getResponseEditTokenData(updateToken, viewAccessLevel)
			if err != nil {
				handleError("invalid view token: "+err.Error(), http.StatusUnauthorized, response)
				return
			}
			if tokenResponseIDString != postid {
				handleError("view token for wrong response: "+err.Error(), http.StatusUnauthorized, response)
				return
			}
		}
	}
	var filepath string
	var fileIndex string
	if posttype == formType {
		fileIndex = formFileIndex
	} else if posttype == responseType {
		fileIndex = responseFileIndex
	} else {
		fileIndex = blogFileIndex
	}
	var addPlaceholderPath string
	if fileType == "image/gif" {
		addPlaceholderPath = placeholderPath
	}
	switch requestType {
	case "original":
		filepath = fileIndex + "/" + postid + "/" + fileid + originalPath
		break
	case "placeholder":
		filepath = fileIndex + "/" + postid + "/" + fileid + addPlaceholderPath + originalPath
		break
	case "blur":
		filepath = fileIndex + "/" + postid + "/" + fileid + addPlaceholderPath + blurPath
		break
	}
	if requestType == "file" {
		var fileobj *storage.ObjectHandle
		fileobj = storageBucket.Object(filepath)
		filereader, err := fileobj.NewReader(ctxStorage)
		if err != nil {
			handleError("error reading file: "+err.Error(), http.StatusBadRequest, response)
			return
		}
		defer filereader.Close()
		filebuffer := new(bytes.Buffer)
		if bytesread, err := filebuffer.ReadFrom(filereader); err != nil {
			handleError("error reading to buffer: num bytes: "+strconv.FormatInt(bytesread, 10)+", "+err.Error(), http.StatusBadRequest, response)
			return
		}
		contentType := filereader.Attrs.ContentType
		response.Header().Set("Content-Type", contentType)
		response.Write(filebuffer.Bytes())
	} else {
		fileURL, err := getSignedURL(filepath, validAccessTypes[2])
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		response.Write([]byte(`{"url":"` + fileURL + `"}`))
	}
}

func getSignedURL(path string, accessType string) (string, error) {
	var accessMethod string
	if accessType == validAccessTypes[2] {
		// view
		accessMethod = "GET"
	} else {
		// edit
		accessMethod = "PUT"
	}
	// form
	return storage.SignedURL(storageBucketName, path, &storage.SignedURLOptions{
		Expires:        time.Now().Add(time.Minute * time.Duration(storageAccessTime)),
		Method:         accessMethod,
		GoogleAccessID: storageAccessID,
		PrivateKey:     storagePrivateKey,
	})
}
