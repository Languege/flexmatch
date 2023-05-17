// redis_wrapper
// @author LanguageY++2013 2022/4/20 5:39 下午
// @company soulgame
package redis

import (
	"errors"
	"fmt"
	"github.com/Languege/flexmatch/common/logger"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"reflect"
	"time"
)

var (
	ErrStreamOutValue = errors.New("stream out must be non-nil pointer to a struct or ptr slice")
	ErrStreamInValue  = errors.New("stream in must be  a struct or struct ptr")
)

type StreamBaseMsg struct {
	ID string `redis:"-"`
}

//ConsumerGroup stream 消费组信息
type ConsumerGroup struct {
	Name            string `redis:"name"`              //消费组名
	Consumers       int    `redis:"consumers"`         //消费组消费者数量
	Pending         int    `redis:"pending"`           //消费组待消费消息数
	LastDeliveredId string `redis:"last-delivered-id"` //最近被传输给消费组的消息ID
	EntriesRead     int    `redis:"entries-read"`      //送给组内消费者的最后一个条目的逻辑 "读取计数器"。
	Lag             int    `redis:"lag"`               //流中仍在等待交付给该组消费者的条目数，如果不能确定该数字，则为NULL。
}

type StreamInfo struct {
	Length            int         `redis:"length"`               //流中的条目数
	RadixTreeKeys     int         `redis:"radix-tree-keys"`      //底层radix数据结构中的键的数量
	RadixTreeNodes    int         `redis:"radix-tree-nodes"`     //底层radix数据结构中的节点数
	Groups            int         `redis:"groups"`               //为该流定义的消费者组的数量
	LastGeneratedId   string      `redis:"last-generated-id"`    //被添加到流中的最近的条目的ID
	MaxDeletedEntryId string      `redis:"max-deleted-entry-id"` //从流中被删除的最大条目ID
	EntriesAdded      int         `redis:"entries-added"`        //在流的生命周期内添加到流中的所有条目的计数
	FirstEntry        StreamEntry `redis:"-"`                    //流中第一个条目的ID和字段值图示
	LastEntry         StreamEntry `redis:"-"`                    //流中最后一个条目的ID和字段值图示
}

//StreamEntry stream 条目
type StreamEntry struct {
	ID     string `redis:"-"`
	Fields map[string]string
}

type ByteStreamEntry struct {
	ID   string `redis:"-"`
	Data []byte `redis:"data"`
}

func(rw *RedisWrapper) XAdd(key string, in interface{}, params ...interface{}) (err error) {
	conn := rw.GetConn()
	defer conn.Close()

	//必须是结构体
	rt := reflect.TypeOf(in)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return ErrStreamInValue
	}
	args := redis.Args{rw.buildKey(key)}
	if len(params) > 0 {
		//设置队列长度
		args = args.Add("MAXLEN", "~", params[0])
	}
	args = args.Add("*")
	args = args.AddFlat(in)

	_, err = conn.Do("XADD", args...)
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XADD"), zap.String("args", fmt.Sprint("%v", args)))
		return
	}
	return
}

func(rw *RedisWrapper) XAddWithMaxLen(key string, in interface{}, maxLen int) (err error) {
	return rw.XAdd(key, in, maxLen)
}

func(rw *RedisWrapper) XLen(key string) (ret int, err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := []interface{}{rw.buildKey(key)}

	ret, err = redis.Int(conn.Do("XLEN", args...))
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XLEN"), zap.String("args", fmt.Sprint("%v", args)))
		return
	}
	return
}

func(rw *RedisWrapper) XRead(key string, count int, timeout time.Duration, out interface{}) (err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := redis.Args{}
	if count > 0 {
		args = args.Add("COUNT", count)
	}

	if timeout > 0 {
		args = args.Add("BLOCK", int(timeout/time.Millisecond))
	}

	args = append(args, "STREAMS",  rw.buildKey(key), "$")

	var reply interface{}
	reply, err = conn.Do("XREAD", args...)
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XREAD"), zap.String("args", fmt.Sprint("%v", args)))
		return
	}

	return streamScanStructSlice(reply, out)
}


func (rw *RedisWrapper) XGroupCreateFromTail(key, groupName string) (reply string, err error) {
	return rw.XGroupCreate(key, groupName, "$")
}

func (rw *RedisWrapper) XGroupCreateFromBeginning(key, groupName string) (reply string, err error) {
	return rw.XGroupCreate(key, groupName, "0-0")
}

// XGroupCreate 创建消费组
//				返回值：
//					reply: 成功返回OK
func (rw *RedisWrapper) XGroupCreate(key, groupName, initId string) (reply string, err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := redis.Args{}
	args = args.Add("CREATE", key, groupName, initId)

	reply, err = redis.String(conn.Do("XGROUP", args...))
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XGROUP"), zap.String("args", fmt.Sprintf("%v", args)))
		return
	}
	return
}

func (rw *RedisWrapper) XGroupDestroy(key, groupName string) (reply string, err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := redis.Args{}
	args = args.Add("DESTROY", key, groupName)

	reply, err = redis.String(conn.Do("XGROUP", args...))
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XGROUP"), zap.String("args", fmt.Sprint("%v", args)))
		return
	}
	return
}

func (rw *RedisWrapper) XReadGroup(group, consumer string, count int, timeout time.Duration, queueKey string, out interface{}) (err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := redis.Args{}
	args = args.Add("GROUP", group, consumer)
	if count > 0 {
		args = args.Add("COUNT", count)
	}

	if timeout > 0 {
		args = args.Add("BLOCK", int(timeout/time.Millisecond))
	}

	args = args.Add("STREAMS", queueKey, ">")
	var reply interface{}
	reply, err = conn.Do("XREADGROUP", args...)
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XREADGROUP"), zap.String("args", fmt.Sprint("%v", args)))
		return
	}

	return streamScanStructSlice(reply, out)
}

func (rw *RedisWrapper) XAck(group, queueKey, id string) (acknowledged int, err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := redis.Args{group, queueKey, id}
	acknowledged, err = redis.Int(conn.Do("XACK", args...))
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XACK"), zap.String("args", fmt.Sprint("%v", args)))
	}

	return
}

//XInfoGroups 返回queueKey关联的消费组列表信息
func (rw *RedisWrapper) XInfoGroups(queueKey string) (ret []*ConsumerGroup, err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := redis.Args{"GROUPS", queueKey}
	var reply interface{}
	reply, err = conn.Do("XINFO", args...)
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XINFO"), zap.String("args", fmt.Sprint("%v", args)))
		return
	}

	for _, v := range reply.([]interface{}) {
		group := &ConsumerGroup{}

		err = redis.ScanStruct(v.([]interface{}), group)
		if err != nil {
			return
		}

		ret = append(ret, group)
	}

	return
}

//XInfoStream  返回存储在queueKey的流基本信息
func (rw *RedisWrapper) XInfoStream(queueKey string) (ret *StreamInfo, err error) {
	conn := rw.GetConn()
	defer conn.Close()

	args := redis.Args{"STREAM", queueKey}
	var reply interface{}
	reply, err = conn.Do("XINFO", args...)
	if err != nil {
		logger.Errorw(err.Error(), zap.String("command", "XINFO"), zap.String("args", fmt.Sprint("%v", args)))
		return
	}

	data := reply.([]interface{})
	ret = &StreamInfo{}
	err = redis.ScanStruct(data, ret)
	if err != nil {
		return
	}

	for i := 0; i < len(data); i++ {
		v := data[i]
		if v, verr := redis.String(v, nil); verr == nil && v == "first-entry" {
			entryData := data[i+1].([]interface{})
			id, _ := redis.String(entryData[0], nil)
			ret.FirstEntry = StreamEntry{
				ID: id,
			}
			ret.FirstEntry.Fields, _ = redis.StringMap(entryData[1].([]interface{}), nil)
			i++
			continue
		}

		if v, verr := redis.String(v, nil); verr == nil && v == "last-entry" {
			entryData := data[i+1].([]interface{})
			id, _ := redis.String(entryData[0], nil)
			ret.LastEntry = StreamEntry{
				ID: id,
			}
			ret.LastEntry.Fields, _ = redis.StringMap(entryData[1].([]interface{}), nil)
			i++
			continue
		}
	}

	return
}

func streamScanStructSlice(reply interface{}, out interface{}) (err error) {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr {
		return ErrStreamOutValue
	}
	indirectValue := rv.Elem()
	if indirectValue.Kind() != reflect.Slice {
		return ErrStreamOutValue
	}

	eleType := indirectValue.Type().Elem()
	indirectEleType := eleType
	if indirectEleType.Kind() == reflect.Ptr {
		indirectEleType = indirectEleType.Elem()
	}
	//元素类型判断
	if indirectEleType.Kind() != reflect.Struct {
		return ErrStreamOutValue
	}

	for _, v := range reply.([]interface{}) {
		queue := v.([]interface{})
		data := queue[1].([]interface{})

		indirectValue = reflect.MakeSlice(indirectValue.Type(), len(data), len(data))
		for i, ele := range data {
			fieldAndArgs := ele.([]interface{})[1].([]interface{})

			indirectEleValue := reflect.New(indirectEleType)
			err = redis.ScanStruct(fieldAndArgs, indirectEleValue.Interface())
			if err != nil {
				return
			}
			//设置消息ID
			if id, idErr := redis.String(ele.([]interface{})[0], nil); idErr == nil {
				indirectEleValue.Elem().FieldByName("ID").Set(reflect.ValueOf(id))
			}

			if eleType.Kind() != reflect.Ptr {
				indirectEleValue = indirectEleValue.Elem()
			}

			indirectValue.Index(i).Set(indirectEleValue)
		}
		break
	}

	rv.Elem().Set(indirectValue)
	return
}
