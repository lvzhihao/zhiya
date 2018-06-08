package uchat

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lvzhihao/zhiya/models"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/spf13/viper"
)

var (
	DefaultChatRoomHeadImageBackgroundColor color.RGBA = color.RGBA{200, 200, 200, 0}
	DefaultChatRoomHeadSizeUnit             int        = 132
	DefaultChatRoomHeadPadding              int        = 3
)

func UpdateChatRoomHeadImage(db *gorm.DB, chatRoomSerialNo string) (*models.ChatRoom, error) {
	ir, err := GetChatRoomHeadImage(db, chatRoomSerialNo, 0)
	if err != nil {
		return nil, err
	}
	var chatRoom models.ChatRoom
	err = db.Where("chat_room_serial_no = ?", chatRoomSerialNo).First(&chatRoom).Error
	if err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("chatroom/headimage/%x-%d.jpeg", md5.Sum([]byte(chatRoomSerialNo)), time.Now().UnixNano())
	data, _ := ioutil.ReadAll(ir)
	ret, err := UploadQiuniu(filename, data)
	if err != nil {
		return nil, err
	}
	chatRoom.HeadImage = fmt.Sprintf("https://%s/%s", viper.GetString("qiniu_zhiya_domain"), ret.Key)
	err = db.Save(&chatRoom).Error
	return &chatRoom, err
}

func GetChatRoomHeadImage(db *gorm.DB, chatRoomSerialNo string, limit int) (io.Reader, error) {
	obj, err := NewChatRoomHeadImage(db, chatRoomSerialNo)
	if err != nil {
		return nil, err
	}
	return obj.Generate(limit)
}

type ChatRoomHeadImage struct {
	db               *gorm.DB
	chatRoomSerialNo string
	members          []models.ChatRoomMember
}

func NewChatRoomHeadImage(db *gorm.DB, chatRoomSerialNo string) (*ChatRoomHeadImage, error) {
	obj := &ChatRoomHeadImage{
		db:               db,
		chatRoomSerialNo: chatRoomSerialNo,
		members:          make([]models.ChatRoomMember, 0),
	}
	err := obj.init()
	return obj, err
}

func (c *ChatRoomHeadImage) init() error {
	err := c.db.Where("chat_room_serial_no = ?", c.chatRoomSerialNo).Where("is_active = ?", true).Limit(20).Order("created_at ASC").Find(&c.members).Error
	return err
}

func (c *ChatRoomHeadImage) fetchImage() []string {
	rst := make([]string, 0)
	for _, member := range c.members {
		rst = append(rst, member.HeadImages)
	}
	return rst
}

func (c *ChatRoomHeadImage) Generate(limit int) (io.Reader, error) {
	images := c.fetchImage()
	obj := NewHeadImage()
	err := obj.SetImages(images, limit)
	if err != nil {
		return nil, err
	}
	return obj.JPEG()
}

// 微信头像生成
type HeadImage struct {
	size    int           // px
	padding int           // padding
	images  []image.Image // images
	board   *image.RGBA   // board
}

func NewHeadImage() *HeadImage {
	return &HeadImage{
		size:    DefaultChatRoomHeadSizeUnit, //default
		padding: DefaultChatRoomHeadPadding,  //default
		images:  make([]image.Image, 0),
	}
}

// 设置画布大小
func (c *HeadImage) SetSize(size int) {
	c.size = size
}

// 设备头像
func (c *HeadImage) SetImages(images []string, limit int) error {
	c.images = make([]image.Image, 0)
	for _, img := range images {
		if limit > 0 && len(c.images) == limit {
			break
		}
		err := c.AddImage(img)
		if err != nil {
			return err
		}
	}
	return nil
}

// 添加一个头像
func (c *HeadImage) AddImage(imgUrl string) error {
	resp, err := http.Get(imgUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		obj, err := jpeg.Decode(resp.Body)
		if err != nil {
			// continue
		} else {
			c.images = append(c.images, obj)
		}
		return nil
	} else {
		return fmt.Errorf("resp error: %s", resp.Status)
	}
}

// 生成
func (c *HeadImage) JPEG() (io.Reader, error) {
	img, err := c.generate()
	if err != nil {
		return nil, err
	} else {
		return c.jpegRender(img)
	}
}

// generate
func (c *HeadImage) generate() (image.Image, error) {
	// write images
	rectangle, err := c.writeImages()
	if err != nil {
		return nil, err
	} else {
		return c.board.SubImage(rectangle), nil
	}
}

func (c *HeadImage) jpegRender(img image.Image) (io.Reader, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	err := jpeg.Encode(writer, img, nil)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b.Bytes()), nil
}

// create board
func (c *HeadImage) createBoard(rectangle image.Rectangle) {
	c.board = image.NewRGBA(rectangle)
	c.backgroundColor(DefaultChatRoomHeadImageBackgroundColor)
}

// fix board background color
func (c *HeadImage) backgroundColor(back color.RGBA) {
	bounds := c.board.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c.board.Set(x, y, back)
		}
	}
}

// wirte image
func (c *HeadImage) writeImages() (image.Rectangle, error) {
	switch len(c.images) {
	case 0:
		return image.Rectangle{}, fmt.Errorf("images nums error")
	case 1:
		return image.Rectangle{}, fmt.Errorf("images nums error")
	case 2:
		return image.Rectangle{}, fmt.Errorf("images nums error")
	case 3:
		rectangle := c.rectangle_2()
		c.writeImage_3()
		return rectangle, nil
	case 4:
		rectangle := c.rectangle_2()
		c.writeImage_4()
		return rectangle, nil
	case 5:
		rectangle := c.rectangle_3()
		c.writeImage_5()
		return rectangle, nil
	case 6:
		rectangle := c.rectangle_3()
		c.writeImage_6()
		return rectangle, nil
	case 7:
		rectangle := c.rectangle_3()
		c.writeImage_7()
		return rectangle, nil
	case 8:
		rectangle := c.rectangle_3()
		c.writeImage_8()
		return rectangle, nil
	case 9:
		rectangle := c.rectangle_3()
		c.writeImage_9()
		return rectangle, nil
	default:
		rectangle := c.rectangle_3()
		c.writeImage_9()
		return rectangle, nil
	}
}

func (c *HeadImage) rectangle_2() image.Rectangle {
	rectangle := image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{c.size*2 + 3*c.padding, c.size*2 + 3*c.padding},
	}
	c.createBoard(rectangle)
	return rectangle
}

func (c *HeadImage) rectangle_3() image.Rectangle {
	rectangle := image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{c.size*3 + 4*c.padding, c.size*3 + 4*c.padding},
	}
	c.createBoard(rectangle)
	return rectangle
}

func (c *HeadImage) writeImage_3() {
	// write
	for i, img := range c.images[0:3] {
		bounds := img.Bounds()
		var xP, yP int
		if i == 0 {
			xP = i%2*(c.size+c.padding) + c.padding + c.size/2
			yP = (i/2)*(c.size+c.padding) + c.padding
		} else {
			xP = (i+1)%2*(c.size+c.padding) + c.padding
			yP = ((i+1)/2)*(c.size+c.padding) + c.padding
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.board.Set(x+xP, y+yP, img.At(x, y))
			}
		}
	}
}

func (c *HeadImage) writeImage_4() {
	// write
	for i, img := range c.images[0:4] {
		bounds := img.Bounds()
		xP := i%2*(c.size+c.padding) + c.padding
		yP := (i/2)*(c.size+c.padding) + c.padding
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.board.Set(x+xP, y+yP, img.At(x, y))
			}
		}
	}
}

func (c *HeadImage) writeImage_5() {
	// write
	for i, img := range c.images[0:5] {
		bounds := img.Bounds()
		var xP, yP int
		if i < 2 {
			xP = i%3*(c.size+c.padding) + c.padding + c.size/2
			yP = (i/3)*(c.size+c.padding) + c.padding + c.size/2
		} else {
			xP = (i+1)%3*(c.size+c.padding) + c.padding
			yP = ((i+1)/3)*(c.size+c.padding) + c.padding + c.size/2
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.board.Set(x+xP, y+yP, img.At(x, y))
			}
		}
	}
}

func (c *HeadImage) writeImage_6() {
	// write
	for i, img := range c.images[0:6] {
		bounds := img.Bounds()
		var xP, yP int
		xP = i%3*(c.size+c.padding) + c.padding
		yP = (i/3)*(c.size+c.padding) + c.padding + c.size/2
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.board.Set(x+xP, y+yP, img.At(x, y))
			}
		}
	}
}

func (c *HeadImage) writeImage_7() {
	// write
	for i, img := range c.images[0:7] {
		bounds := img.Bounds()
		var xP, yP int
		if i < 1 {
			xP = i%3*(c.size+c.padding) + c.padding + c.size
			yP = (i/3)*(c.size+c.padding) + c.padding
		} else {
			xP = (i+2)%3*(c.size+c.padding) + c.padding
			yP = ((i+2)/3)*(c.size+c.padding) + c.padding
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.board.Set(x+xP, y+yP, img.At(x, y))
			}
		}
	}
}

func (c *HeadImage) writeImage_8() {
	// write
	for i, img := range c.images[0:8] {
		bounds := img.Bounds()
		var xP, yP int
		if i < 2 {
			xP = i%3*(c.size+c.padding) + c.padding + c.size/2
			yP = (i/3)*(c.size+c.padding) + c.padding
		} else {
			xP = (i+1)%3*(c.size+c.padding) + c.padding
			yP = ((i+1)/3)*(c.size+c.padding) + c.padding
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.board.Set(x+xP, y+yP, img.At(x, y))
			}
		}
	}
}

func (c *HeadImage) writeImage_9() {
	// write
	for i, img := range c.images[0:9] {
		bounds := img.Bounds()
		xP := i%3*(c.size+c.padding) + c.padding
		yP := (i/3)*(c.size+c.padding) + c.padding
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				c.board.Set(x+xP, y+yP, img.At(x, y))
			}
		}
	}
}

func UploadQiuniu(filename string, data []byte) (storage.PutRet, error) {
	putPolicy := storage.PutPolicy{
		Scope: viper.GetString("qiniu_zhiya_bucket"),
	}
	mac := qbox.NewMac(viper.GetString("qiniu_access_key"), viper.GetString("qiniu_secret_key"))
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
	/*
		Params: map[string]string{
			"x:name": "",
		},
	*/
	}
	dataLen := int64(len(data))
	err := formUploader.Put(context.Background(), &ret, upToken, filename, bytes.NewReader(data), dataLen, &putExtra)
	return ret, err
}
