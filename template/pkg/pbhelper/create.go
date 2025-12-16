package pbhelper

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// CreateProtoMessageByName **根据 "proto包名.消息名" 反射创建 proto.Message 对象**
func CreateProtoMessageByName(messageName string) (proto.Message, error) {
	// 从全局注册表中查找类型
	mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(messageName))
	if err != nil {
		return nil, fmt.Errorf("消息类型未找到: %s", messageName)
	}
	return mt.New().Interface(), nil
}
