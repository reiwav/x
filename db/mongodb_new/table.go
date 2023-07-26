package mongodb_new

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Table struct {
	*mongo.Collection
	Prefix string
}

func NewTable(name, prefix string, db *mongo.Database) *Table {
	fmt.Println("DB", name)
	return &Table{
		Collection: db.Collection(name),
		Prefix:     prefix,
	}
}

func (t *Table) CreateIndexOne(mode mongo.IndexModel, opts ...*options.CreateIndexesOptions) error {
	ctx := context.Background()
	var _, err = t.Indexes().CreateOne(ctx, mode, opts...)
	return err
}

func (t *Table) CreateIndexMany(mods []mongo.IndexModel, opts ...*options.CreateIndexesOptions) error {
	ctx := context.Background()
	var _, err = t.Indexes().CreateMany(ctx, mods, opts...)
	return err
}

func (t *Table) Create(model IModel) error {
	ctx := context.Background()
	model.BeforeCreate(t.Prefix)
	var _, err = t.InsertOne(ctx, model)
	if err != nil {
		logDB.Errorf("Create table "+t.Name()+": "+err.Error(), model)
	}
	return err
}

func (t *Table) CreateForce(model IModel) error {
	ctx := context.Background()
	var _, err = t.InsertOne(ctx, model)
	if err != nil {
		logDB.Errorf("Create table "+t.Name()+": "+err.Error(), model)
	}
	return err
}

func (t *Table) Update(id string, model IModel) error {
	ctx := context.Background()
	model.BeforeUpdate()
	var _, err = t.UpdateOne(ctx, bson.M{"_id": id, "deleted_at": 0}, bson.M{"$set": model})
	if err != nil {
		logDB.Errorf("Update table "+t.Name()+": "+err.Error(), model)
	}
	return err
}

func (t *Table) Delete(id string, model IModel) error {
	ctx := context.Background()
	model.BeforeDelete()
	var _, err = t.UpdateOne(ctx, bson.M{"_id": id, "deleted_at": 0}, bson.M{"$set": bson.M{"deleted_at": time.Now().Unix()}})
	if err != nil {
		logDB.Errorf("Delete table "+t.Name()+": "+err.Error(), model)
	}
	return err
}

func (t *Table) DeleteByID(id string) error {
	var err = t.UnsafeUpdateByID(id, bson.M{"deleted_at": time.Now().Unix()})
	if err != nil {
		logDB.Errorf("DeleteByID "+err.Error(), id)
	}
	return err
}
func (t *Table) SelectAndDelete(id string) error {
	ctx := context.Background()
	var timeNow = time.Now().Unix()
	after := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	var res = t.FindOneAndUpdate(ctx, bson.M{"_id": id, "deleted_at": 0},
		bson.M{"$set": bson.M{"deleted_at": timeNow, "created_at": timeNow}}, opts)
	if res.Err() != nil {
		logDB.Errorf("Delete table " + t.Name() + ": " + res.Err().Error())
	}
	return res.Err()
}

func (t *Table) UnsafeUpdate(filter bson.M, v interface{}) error {
	ctx := context.Background()
	filter["deleted_at"] = 0
	var _, err = t.UpdateOne(ctx, filter,
		bson.M{"$set": v})
	if err != nil {
		logDB.Errorf("UnsafeUpdate table "+t.Name()+": "+err.Error(), v)
	}
	return err
}

func (t *Table) UpdateForce(filter bson.M, v interface{}) error {
	ctx := context.Background()
	filter["deleted_at"] = 0
	var _, err = t.UpdateOne(ctx, filter, v)
	if err != nil {
		logDB.Errorf("UnsafeUpdate table "+t.Name()+": "+err.Error(), v)
	}
	return err
}

func (t *Table) UnsafeUpdateByID(id string, v interface{}) error {
	ctx := context.Background()
	var _, err = t.UpdateOne(ctx,
		bson.M{"deleted_at": 0,
			"_id": id},
		bson.M{"$set": v})
	if err != nil {
		logDB.Errorf("UnsafeUpdateByID table "+t.Name()+": "+err.Error(), v)
	}
	return err
}

func (t *Table) CreateMany(v []interface{}) ([]interface{}, error) {
	ctx := context.Background()
	var res, err = t.InsertMany(ctx, v)
	var ids []interface{}
	if err != nil {
		logDB.Errorf("UnsafeUpdateByID table "+t.Name()+": "+err.Error(), v)
	}
	if res != nil {
		ids = res.InsertedIDs
	}
	return ids, err
}

func (t *Table) SelectOne(filter bson.M, v interface{}) error {
	ctx := context.Background()
	filter["deleted_at"] = 0
	var err = t.FindOne(ctx, filter).Decode(v)
	return err
}

func (t *Table) SelectOneWithFields(filter bson.M, v interface{}, fields bson.M) error {
	ctx := context.Background()
	filter["deleted_at"] = 0

	var opts = options.FindOne().SetProjection(fields)
	var err = t.FindOne(ctx, filter, opts).Decode(v)
	return err
}

func (t *Table) SelectManyWithFields(filter bson.M, v interface{}, fields bson.M) error {
	ctx := context.Background()
	filter["deleted_at"] = 0

	var opts = options.Find().SetProjection(fields)
	var cur, err = t.Find(ctx, filter, opts)
	if err != nil {
		cur.Close(ctx)
		return err
	}
	err = cur.All(ctx, v)
	return err
}

func (t *Table) SelectByID(id string, v interface{}) error {
	ctx := context.Background()
	var filter = bson.M{
		"deleted_at": 0,
		"_id":        id,
	}
	var err = t.FindOne(ctx, filter).Decode(v)

	return err
}

func (t *Table) SelectMany(filter bson.M, v interface{}) error {
	ctx := context.Background()
	filter["deleted_at"] = 0
	var cur, err = t.Find(ctx, filter)
	if err != nil {
		cur.Close(ctx)
		return err
	}
	err = cur.All(ctx, v)
	return err
}

func (t *Table) SelectDistinct(field string, filter bson.M) ([]interface{}, error) {
	ctx := context.Background()
	filter["deleted_at"] = 0
	return t.Distinct(ctx, field, filter)
}

func (t *Table) UpdateAll(filter bson.M, update interface{}) error {
	ctx := context.Background()
	filter["deleted_at"] = 0
	var _, err = t.UpdateMany(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return err
	}
	return err
}

func (t *Table) SelectAndSort(filter bson.M, sortFields []string, skip, limit int64, res interface{}) error {
	ctx := context.Background()
	filter["deleted_at"] = 0
	var opts = options.Find()
	if len(sortFields) > 0 {
		sortBson := bson.D{}
		for _, val := range sortFields {
			if strings.Contains(val, "-") {
				item := strings.Split(val, "-")[1]
				sortBson = append(sortBson, bson.E{item, -1})
			} else {
				sortBson = append(sortBson, bson.E{val, 1})
			}
		}
		opts.SetSort(sortBson)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	var cur, err = t.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	err = cur.All(ctx, res)
	return err
}

func (t *Table) Pipe(pipeline []bson.M, res interface{}) error {
	ctx := context.Background()
	var cur, err = t.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	err = cur.All(ctx, res)
	return err
}

func (t *Table) Count(filter bson.M) (int64, error) {
	filter["deleted_at"] = 0
	ctx := context.Background()
	return t.CountDocuments(ctx, filter)
}
