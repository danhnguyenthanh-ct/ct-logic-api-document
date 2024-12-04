package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/carousell/ct-go/pkg/container"

	"github.com/carousell/ct-core-uni-free-premium-service/config"
)

func TestBaseCollection_InsertOne(t *testing.T) {
	userCol := getUserCollection((t))
	user := randomUser()

	ctx := context.Background()
	err := userCol.Insert(ctx, user)
	require.Nil(t, err)
	require.NotEmpty(t, user.Id)
	filter := primitive.M{
		"_id":        user.Id,
		"account_id": user.AccountId,
	}
	found, err := userCol.Get(ctx, filter)
	require.Nil(t, err)
	require.NotNil(t, found)
	require.True(t, cmp.Equal(user, found,
		cmp.AllowUnexported(testUser{}),
		cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
	require.WithinDuration(t, time.Now(), *found.CreatedAt, time.Minute)
	require.WithinDuration(t, time.Now(), *found.UpdatedAt, time.Minute)
}

func TestBaseCollection_Get(t *testing.T) {
	t.Parallel()
	userCol := getUserCollection((t))
	user := seedUser(t, userCol)

	testCases := []struct {
		Name   string
		Found  bool
		Params container.Map
	}{
		{"success", true, container.Map{"age": user.Age, "name": user.Name, "account_id": user.AccountId}},
		{"not_found", false, container.Map{"age": "invalid", "name": "invalid", "account_id": gofakeit.IntRange(1, 10000000)}},
	}

	for _, testCase := range testCases {
		tc := testCase

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			found, err := userCol.Get(ctx, tc.Params)
			require.Nil(t, err)
			if tc.Found {
				require.NotNil(t, found)
				require.True(t, cmp.Equal(user, found,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
			} else {
				require.Nil(t, found)
			}
		})
	}
}

func TestBaseCollection_GetByBatch(t *testing.T) {
	t.Parallel()
	userCol := getUserCollection((t))
	user := seedUser(t, userCol)

	testCases := []struct {
		Name   string
		Found  bool
		Params container.Map
	}{
		{"success", true, container.Map{"age": user.Age, "name": user.Name, "account_id": user.AccountId}},
		{"not_found", false, container.Map{"age": "invalid", "name": "invalid", "account_id": user.AccountId}},
		{"query_without_account_id", true, container.Map{"age": user.Age, "name": user.Name}},
	}

	for _, testCase := range testCases {
		tc := testCase

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			found, err := userCol.GetByBatch(ctx, tc.Params, primitive.D{}, 100, 0)
			require.Nil(t, err)
			if tc.Found {
				require.True(t, len(found) > 0)
				require.True(t, cmp.Equal(user, found[0],
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
			} else {
				require.Empty(t, found)
			}
		})
	}
}

func TestBaseCollection_UpdateById(t *testing.T) {
	t.Parallel()

	userCol := getUserCollection((t))
	user := seedUser(t, userCol)
	originalId := user.Id
	newId := randomObjectID()

	// update values
	user.Name = gofakeit.Name()
	user.Age = gofakeit.IntRange(1, 100)

	time.Sleep(time.Second) // to have different UpdatedAt

	testCases := []struct {
		Name    string
		Id      primitive.ObjectID
		Success bool
		Param   container.Map
	}{
		{"success", originalId, true, container.Map{"_id": originalId, "account_id": user.AccountId}},
		{"not_found", newId, false, container.Map{"_id": newId, "account_id": user.AccountId}},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			modifiedCount, err := userCol.Update(context.Background(), tc.Param, user)
			require.Nil(t, err)
			updated, err := userCol.Get(context.Background(), tc.Param)
			require.Nil(t, err)
			if tc.Success {
				require.Equal(t, int64(1), modifiedCount)
				require.True(t, cmp.Equal(user, updated,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
				require.True(t, updated.UpdatedAt.Sub(*user.UpdatedAt).Seconds() >= 1)
			} else {
				require.Equal(t, int64(0), modifiedCount)
				require.Equal(t, user.Id, originalId) // no upsert -> no Id update
				require.Nil(t, updated)
			}
		})
	}
}

func TestBaseCollection_Update(t *testing.T) {
	t.Parallel()
	userCol := getUserCollection((t))
	user := seedUser(t, userCol)

	// update values
	user.Age = gofakeit.IntRange(1, 100)

	time.Sleep(time.Second) // to have different UpdatedAt

	testCases := []struct {
		Name         string
		UpdateParams container.Map
		Success      bool
		GetParams    container.Map
	}{
		{
			"success",
			container.Map{"name": user.Name, "account_id": user.AccountId},
			true,
			container.Map{"account_id": user.AccountId, "_id": user.Id},
		},
		{
			"not_found",
			container.Map{"name": "invalid", "account_id": user.AccountId},
			false,
			container.Map{"account_id": user.AccountId, "_id": user.Id},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			modifiedCount, err := userCol.Update(context.Background(), tc.UpdateParams, user)
			require.Nil(t, err)
			if tc.Success {
				require.Equal(t, int64(1), modifiedCount)
				updated, err := userCol.Get(context.Background(), tc.GetParams)
				require.Nil(t, err)
				require.NotNil(t, updated)
				require.True(t, cmp.Equal(user, updated,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
				require.WithinDuration(t, *updated.UpdatedAt, *user.UpdatedAt, time.Minute)
			} else {
				require.Equal(t, int64(0), modifiedCount)
			}
		})
	}
}

func TestBaseCollection_UpdatePartial(t *testing.T) {
	t.Parallel()
	userCol := getUserCollection((t))
	user := seedUser(t, userCol)

	// update values
	user.Age = gofakeit.IntRange(1, 100)

	time.Sleep(time.Second) // to have different UpdatedAt

	testCases := []struct {
		Name                string
		UpdatePartialParams container.Map
		Success             bool
		GetParams           container.Map
	}{
		{
			"success",
			container.Map{"name": user.Name, "account_id": user.AccountId},
			true,
			container.Map{"account_id": user.AccountId, "_id": user.Id},
		},
		{
			"not_found",
			container.Map{"name": "invalid", "account_id": user.AccountId},
			false,
			container.Map{"account_id": user.AccountId, "_id": user.Id},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			modifiedCount, err := userCol.UpdatePartial(context.Background(), tc.UpdatePartialParams, container.Map{
				"age": user.Age,
			})
			require.Nil(t, err)
			if tc.Success {
				require.Equal(t, int64(1), modifiedCount)
				updated, err := userCol.Get(context.Background(), tc.GetParams)
				require.Nil(t, err)
				require.NotNil(t, updated)
				require.True(t, cmp.Equal(user, updated,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
				require.True(t, updated.UpdatedAt.Sub(*user.UpdatedAt).Seconds() >= 1)
			} else {
				require.Equal(t, int64(0), modifiedCount)
			}
		})
	}
}

func TestBaseCollection_Upsert(t *testing.T) {
	t.Parallel()
	userCol := getUserCollection((t))
	modifiedUser := seedUser(t, userCol)
	upsertedUser := seedUser(t, userCol)
	originalId := upsertedUser.Id

	// update values
	upsertedUser.Age = gofakeit.IntRange(1, 100)
	upsertedUser.AccountId = gofakeit.IntRange(1, 10000000)
	upsertedUser.Name = gofakeit.Name()
	modifiedUser.Age = gofakeit.IntRange(1, 100)

	time.Sleep(time.Second) // to have different UpdatedAt

	testCases := []struct {
		Name         string
		UpsertParams container.Map
		User         *testUser
		IsUpsert     bool
	}{
		{"success_modified", container.Map{"name": modifiedUser.Name, "account_id": modifiedUser.AccountId}, modifiedUser, false},
		{"success_upserted", container.Map{"name": upsertedUser.Name, "account_id": upsertedUser.AccountId}, upsertedUser, true},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			modifiedCount, upsertedCount, err := userCol.Upsert(context.Background(), tc.UpsertParams, tc.User)
			require.Nil(t, err)
			updated, err := userCol.Get(context.Background(), container.Map{"account_id": tc.User.AccountId})
			require.Nil(t, err)
			require.NotNil(t, updated)

			if tc.IsUpsert {
				require.Equal(t, int64(0), modifiedCount)
				require.Equal(t, int64(1), upsertedCount)
				require.True(t, cmp.Equal(tc.User, updated,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
				require.True(t, updated.UpdatedAt.Sub(*tc.User.UpdatedAt).Seconds() >= 1)
				require.NotEqual(t, originalId, updated.Id)
			} else {
				require.Equal(t, int64(1), modifiedCount)
				require.Equal(t, int64(0), upsertedCount)
				require.True(t, cmp.Equal(tc.User, updated,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
			}
		})
	}
}

func TestBaseCollection_UpsertById(t *testing.T) {
	t.Parallel()
	userCol := getUserCollection((t))
	modifiedUser := seedUser(t, userCol)
	upsertedUser := seedUser(t, userCol)
	originalId := upsertedUser.Id

	// update values
	upsertedUser.Age = gofakeit.IntRange(1, 100)
	modifiedUser.Age = gofakeit.IntRange(1, 100)

	time.Sleep(time.Second) // to have different UpdatedAt

	testCases := []struct {
		Name     string
		Id       primitive.ObjectID
		User     *testUser
		IsUpsert bool
		Param    container.Map
	}{
		{
			"success_modified",
			modifiedUser.Id,
			modifiedUser,
			false,
			container.Map{"_id": modifiedUser.Id, "account_id": modifiedUser.AccountId},
		},
		{
			"success_upserted",
			randomObjectID(),
			upsertedUser,
			true,
			container.Map{"_id": randomObjectID(), "account_id": upsertedUser.AccountId},
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			modifiedCount, upsertedCount, err := userCol.Upsert(context.Background(), tc.Param, tc.User)
			require.Nil(t, err)
			updated, err := userCol.Get(context.Background(), tc.Param)
			require.Nil(t, err)
			require.NotNil(t, updated)

			if tc.IsUpsert {
				require.Equal(t, int64(0), modifiedCount)
				require.Equal(t, int64(1), upsertedCount)
				require.True(t, cmp.Equal(tc.User, updated,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
				require.True(t, updated.UpdatedAt.Sub(*tc.User.UpdatedAt).Seconds() >= 1)
				require.NotEqual(t, originalId, updated.Id) // upsert, so different Id
			} else {
				require.Equal(t, int64(1), modifiedCount)
				require.Equal(t, int64(0), upsertedCount)
				require.True(t, cmp.Equal(tc.User, updated,
					cmp.AllowUnexported(testUser{}),
					cmpopts.IgnoreFields(testUser{}, "BaseEntity.UpdatedAt", "BaseEntity.CreatedAt")))
			}
		})
	}
}

func TestBaseCollection_DeleteById(t *testing.T) {
	t.Parallel()

	userCol := getUserCollection((t))
	user := seedUser(t, userCol)
	filterFound := primitive.M{
		"_id":        user.Id,
		"account_id": user.AccountId,
	}
	err := userCol.Delete(context.Background(), filterFound, nil)
	require.Nil(t, err)

	// not found case
	filterNotFound := primitive.M{
		"_id":        user.Id,
		"account_id": gofakeit.IntRange(1, 10000000),
	}
	err = userCol.Delete(context.Background(), filterNotFound, nil)
	require.Nil(t, err)
}

type anotherObject struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type testObject struct {
	// BaseEntity       `bson:",inline"` // not supported yet
	AccountId        int64                  `bson:"account_id" json:"account_id"`
	Phone            string                 `bson:"phone" json:"phone"`
	OrderData        container.Map          `bson:"order_data" json:"order_data"`
	LockedAt         time.Time              `bson:"locked_at" json:"locked_at"`
	GoogleProductIds container.List[string] `bson:"google_product_ids" json:"google_product_ids"`
	Acknowledged     bool                   `bson:"acknowledged" json:"acknowledged"`
	AcknowledgedAt   time.Time              `bson:"acknowledged_at" json:"acknowledged_at"`
	Another          anotherObject          `bson:"another" json:"another"`
	Omit             string                 `bson:"omit,omitempty" json:"omit,omitempty"`
}

func TestStructToMap(t *testing.T) {
	t.Parallel()

	obj := testObject{}
	err := gofakeit.Struct(&obj)
	if err != nil {
		t.Fatal(err)
	}

	obj.Omit = ""
	m := structToMap(obj, true)
	require.Equal(t, obj.AccountId, m["account_id"])
	require.Equal(t, obj.Phone, m["phone"])
	require.Equal(t, obj.OrderData, m["order_data"])
	require.Equal(t, obj.LockedAt, m["locked_at"])
	require.Equal(t, obj.GoogleProductIds, m["google_product_ids"])
	require.Equal(t, obj.Acknowledged, m["acknowledged"])
	require.Equal(t, obj.AcknowledgedAt, m["acknowledged_at"])
	require.Equal(t, obj.Another, m["another"])
	require.Equal(t, nil, m["omit"])
}

type testUser struct {
	BaseEntity `bson:",inline"`
	Name       string `json:"name" bson:"name"`
	Age        int    `json:"age" bson:"age"`
	AccountId  int    `json:"account_id" bson:"account_id"`
}

type UserCollection struct {
	BaseCollection[testUser, *testUser]
}

func getUserCollection(t *testing.T) *UserCollection {
	cfg := config.MustLoad()
	cfg.Mongo.Debug = true
	db, err := ConnectDatabase(context.Background(), cfg)
	require.Nil(t, err)
	col := NewBaseCollection[testUser](db, "test_user")
	return &UserCollection{*col}
}

func randomUser() *testUser {
	return &testUser{
		BaseEntity: BaseEntity{},
		Name:       gofakeit.Name(),
		Age:        gofakeit.IntRange(1, 100),
		AccountId:  gofakeit.IntRange(1, 10000000),
	}
}

func seedUser(t *testing.T, col *UserCollection) *testUser {
	user := randomUser()
	ctx := context.Background()
	err := col.Insert(ctx, user)
	require.Nil(t, err)
	require.NotEmpty(t, user.Id)

	return user
}

func randomObjectID() primitive.ObjectID {
	return primitive.NewObjectIDFromTimestamp(time.Now())
}
