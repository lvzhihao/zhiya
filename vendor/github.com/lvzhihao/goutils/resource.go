package goutils

import "reflect"

type Resource struct {
	Name  string
	Value interface{}
	Tags  []*Tag
}

func NewResource(value interface{}) *Resource {
	res := &Resource{
		Name:  ModelType(value).String(),
		Value: value,
	}
	return res
}

func (this *Resource) Tag(tag *Tag) *Tag {
	tag.Resource = this
	if oldTag := this.GetTag(tag.Name); oldTag != nil {
		oldTag = tag
	} else {
		this.Tags = append(this.Tags, tag)
	}
	return tag
}

func (this *Resource) GetResource() *Resource {
	return this
}

func (this *Resource) GetTag(name string) *Tag {
	for _, tag := range this.Tags {
		if tag.Name == name {
			return tag
		}
	}
	return nil
}

func (this *Resource) NewStruct() interface{} {
	if this.Value == nil {
		return nil
	} else {
		return reflect.New(reflect.Indirect(reflect.ValueOf(this.Value)).Type()).Interface()
	}
}
func (this *Resource) NewSlice() interface{} {
	if this.Value == nil {
		return nil
	}
	sliceType := reflect.SliceOf(reflect.TypeOf(this.Value))
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface()
}

func ModelType(value interface{}) reflect.Type {
	reflectType := reflect.Indirect(reflect.ValueOf(value)).Type()
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}
