package codec

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	schemaregistry "github.com/datamountaineer/schema-registry"
	"github.com/linkedin/goavro"
	segmentKafka "github.com/segmentio/kafka-go"
	"reflect"
)

type Codec struct {
	schemaClient *schemaregistry.Client
}

type RecordMessage struct {
	After       map[string]map[string]interface{} `json:"after"`
	Before      map[string]interface{}            `json:"before"`
	Op          string                            `json:"op"`
	Source      map[string]interface{}            `json:"source"`
	Transaction map[string]interface{}            `json:"transaction"`
	TsMs        map[string]interface{}            `json:"ts_ms"`
}

func (c *Codec) GetSchema(schemaId int) (schemaData string, err error) {
	return c.schemaClient.GetSchemaByID(schemaId)
}

func (c *Codec) Decode(message segmentKafka.Message, out interface{}) error {
	if len(message.Value) < 5 {
		return errors.New("invalid message")
	}
	schema, err := c.GetSchema(int(binary.BigEndian.Uint32(message.Value[1:5])))
	if err != nil {
		return err
	}
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return err
	}
	data, _, err := codec.NativeFromBinary(message.Value[5:])
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var dataOut RecordMessage
	err = json.Unmarshal(bytes, &dataOut)
	if err != nil {
		return err
	}
	topicValue := fmt.Sprintf("%s.Value", message.Topic)

	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		return errors.New("struct output is not a ptr")
	}
	for i := 0; i < reflect.TypeOf(out).Elem().NumField(); i++ {
		for key, value := range dataOut.After[topicValue] {

			if reflect.TypeOf(out).Elem().Field(i).Tag.Get("json") != key {
				continue
			}
			switch reflect.ValueOf(value).Kind() {
			case reflect.Map:
				for _, k := range reflect.ValueOf(value).MapKeys() {
					if reflect.ValueOf(value).MapIndex(k).Elem().String() == "NULL::character varying" {
						break
					}
					reflect.ValueOf(out).Elem().Field(i).Set(reflect.ValueOf(value).MapIndex(k).Elem().Convert(reflect.ValueOf(out).Elem().Field(i).Type()))
					break
				}
			case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
				reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				reflect.ValueOf(out).Elem().Field(i).Set(reflect.ValueOf(value).Convert(reflect.ValueOf(out).Elem().Field(i).Type()))

			default:
				continue
			}
		}
	}
	return err
}

func NewCodec(schemaURL string) (*Codec, error) {
	schemaClient, err := schemaregistry.NewClient(schemaURL)
	if err != nil {
		return nil, err
	}
	return &Codec{
		schemaClient: schemaClient,
	}, nil
}
