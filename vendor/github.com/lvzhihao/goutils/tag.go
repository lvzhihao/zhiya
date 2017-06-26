package goutils

//string `json:"product_id" name:"产品标识" attr:"C(100);required;memo(产品标识)"`

type Tag struct {
	Name     string
	Type     string
	Label    string
	Memo     string
	Resource *Resource
}
