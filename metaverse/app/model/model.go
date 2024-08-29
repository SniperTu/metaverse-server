package model

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 数据库公用上下文
var Context context.Context = context.Background()

type CModel interface {
	ToMap(v CModel) map[string]interface{}
	AddMany(data []interface{}) error
	Add(data interface{}) error
	Remove(where CModel) error
	One(where CModel, v interface{}) error
	All(where CModel, v interface{}) error
	Update(where CModel, data CModel) error
}

type Model struct {
	Coll        *mongo.Collection `json:"-" bson:"-"`
	ChangeField []string          `json:"-" bson:"-"`
}

func (this *Model) ToMap(v CModel) map[string]interface{} {
	keys := reflect.TypeOf(v)
	vals := reflect.ValueOf(v)

	if vals.Kind() == reflect.Ptr {
		if vals.IsNil() {
			return nil
		}
		originType := reflect.ValueOf(v).Elem().Type()

		if originType.Kind() != reflect.Struct {
			return nil
		}

		keys = keys.Elem()
		vals = vals.Elem()
	}

	var rs = make(map[string]interface{})

	var length = keys.NumField()
	if length >= 2 {
		keys = vals.Field(0).Type()
		vals = vals.Field(0)
		length = keys.NumField()
		for i := 0; i < length; i++ {
			key := keys.Field(i).Tag.Get("bson")
			key = strings.Split(key, ",")[0]
			val := vals.Field(i).Interface()
			for _, s := range this.ChangeField {
				if s == key {
					rs[key] = val
					break
				}
			}
		}
	}
	return rs
}

// 设置集合
func (this *Model) SetColl(coll *mongo.Collection) CModel {
	this.Coll = coll
	return this
}

// 数据批量入库
func (this *Model) AddMany(data []interface{}) error {
	if this.Coll == nil {
		return errors.New("数据库未连接")
	}
	_, err := this.Coll.InsertMany(context.Background(), data)
	if err != nil {
		return err
	}
	return nil
}

// 数据入库
func (this *Model) Add(data interface{}) error {
	if this.Coll == nil {
		return errors.New("数据库未连接")
	}
	_, err := this.Coll.InsertOne(context.Background(), data)
	if err != nil {
		return err
	}
	return nil
}

// 删除数据
func (this *Model) Remove(where CModel) error {
	if this.Coll == nil {
		return errors.New("数据库未连接")
	}
	filter := where.ToMap(where)
	_, err := this.Coll.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

// 修改数据
func (this *Model) Update(where CModel, data CModel) error {
	if this.Coll == nil {
		return errors.New("数据库未连接")
	}
	filter := where.ToMap(where)
	set := bson.M{"$set": data.ToMap(data)}
	_, err := this.Coll.UpdateMany(context.Background(), filter, set)
	if err != nil {
		return err
	}
	return nil
}

// 查询单个数据
func (this *Model) One(where CModel, v interface{}) error {
	if this.Coll == nil {
		return errors.New("数据库未连接")
	}
	filter := where.ToMap(where)
	err := this.Coll.FindOne(context.Background(), filter).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

// 查询多个数据
func (this *Model) All(where CModel, v interface{}) error {
	if this.Coll == nil {
		return errors.New("数据库未连接")
	}
	filter := where.ToMap(where)
	var opt = new(options.FindOptions)
	query, err := this.Coll.Find(context.Background(), filter, opt.SetSort(bson.M{"_id": -1}))
	if err != nil {
		return err
	}
	err = query.All(context.Background(), v)
	return err
}
