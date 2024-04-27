// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: search/v1/search.proto

package searchv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SearchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Expression string `protobuf:"bytes,1,opt,name=expression,proto3" json:"expression,omitempty"`
	Uid        int64  `protobuf:"varint,2,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *SearchRequest) Reset() {
	*x = SearchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_search_v1_search_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchRequest) ProtoMessage() {}

func (x *SearchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_search_v1_search_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchRequest.ProtoReflect.Descriptor instead.
func (*SearchRequest) Descriptor() ([]byte, []int) {
	return file_search_v1_search_proto_rawDescGZIP(), []int{0}
}

func (x *SearchRequest) GetExpression() string {
	if x != nil {
		return x.Expression
	}
	return ""
}

func (x *SearchRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

type SearchResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	User    *UserResult    `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	Article *ArticleResult `protobuf:"bytes,2,opt,name=article,proto3" json:"article,omitempty"`
}

func (x *SearchResponse) Reset() {
	*x = SearchResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_search_v1_search_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchResponse) ProtoMessage() {}

func (x *SearchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_search_v1_search_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchResponse.ProtoReflect.Descriptor instead.
func (*SearchResponse) Descriptor() ([]byte, []int) {
	return file_search_v1_search_proto_rawDescGZIP(), []int{1}
}

func (x *SearchResponse) GetUser() *UserResult {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *SearchResponse) GetArticle() *ArticleResult {
	if x != nil {
		return x.Article
	}
	return nil
}

type UserResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Users []*User `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
}

func (x *UserResult) Reset() {
	*x = UserResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_search_v1_search_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserResult) ProtoMessage() {}

func (x *UserResult) ProtoReflect() protoreflect.Message {
	mi := &file_search_v1_search_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserResult.ProtoReflect.Descriptor instead.
func (*UserResult) Descriptor() ([]byte, []int) {
	return file_search_v1_search_proto_rawDescGZIP(), []int{2}
}

func (x *UserResult) GetUsers() []*User {
	if x != nil {
		return x.Users
	}
	return nil
}

type ArticleResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Articles []*Article `protobuf:"bytes,1,rep,name=articles,proto3" json:"articles,omitempty"`
}

func (x *ArticleResult) Reset() {
	*x = ArticleResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_search_v1_search_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ArticleResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ArticleResult) ProtoMessage() {}

func (x *ArticleResult) ProtoReflect() protoreflect.Message {
	mi := &file_search_v1_search_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ArticleResult.ProtoReflect.Descriptor instead.
func (*ArticleResult) Descriptor() ([]byte, []int) {
	return file_search_v1_search_proto_rawDescGZIP(), []int{3}
}

func (x *ArticleResult) GetArticles() []*Article {
	if x != nil {
		return x.Articles
	}
	return nil
}

var File_search_v1_search_proto protoreflect.FileDescriptor

var file_search_v1_search_proto_rawDesc = []byte{
	0x0a, 0x16, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x2e, 0x76, 0x31, 0x1a, 0x14, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x73,
	0x79, 0x6e, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x41, 0x0a, 0x0d, 0x53, 0x65, 0x61,
	0x72, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78,
	0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x22, 0x6f, 0x0a, 0x0e,
	0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x29,
	0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x12, 0x32, 0x0a, 0x07, 0x61, 0x72, 0x74,
	0x69, 0x63, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x73, 0x65, 0x61,
	0x72, 0x63, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x52, 0x07, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x22, 0x33, 0x0a,
	0x0a, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x25, 0x0a, 0x05, 0x75,
	0x73, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x65, 0x61,
	0x72, 0x63, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x05, 0x75, 0x73, 0x65,
	0x72, 0x73, 0x22, 0x3f, 0x0a, 0x0d, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x12, 0x2e, 0x0a, 0x08, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2e, 0x76,
	0x31, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x08, 0x61, 0x72, 0x74, 0x69, 0x63,
	0x6c, 0x65, 0x73, 0x32, 0x4e, 0x0a, 0x0d, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x06, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x12, 0x18,
	0x2e, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x73, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x42, 0xa6, 0x01, 0x0a, 0x0d, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x41, 0x6e, 0x77, 0x65, 0x6e, 0x79, 0x61, 0x2f, 0x47, 0x65, 0x65, 0x6b, 0x54, 0x69, 0x6d,
	0x65, 0x2f, 0x77, 0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2f, 0x76, 0x31,
	0x3b, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x53, 0x58, 0x58, 0xaa,
	0x02, 0x09, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x09, 0x53, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x15, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x0a, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_search_v1_search_proto_rawDescOnce sync.Once
	file_search_v1_search_proto_rawDescData = file_search_v1_search_proto_rawDesc
)

func file_search_v1_search_proto_rawDescGZIP() []byte {
	file_search_v1_search_proto_rawDescOnce.Do(func() {
		file_search_v1_search_proto_rawDescData = protoimpl.X.CompressGZIP(file_search_v1_search_proto_rawDescData)
	})
	return file_search_v1_search_proto_rawDescData
}

var file_search_v1_search_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_search_v1_search_proto_goTypes = []interface{}{
	(*SearchRequest)(nil),  // 0: search.v1.SearchRequest
	(*SearchResponse)(nil), // 1: search.v1.SearchResponse
	(*UserResult)(nil),     // 2: search.v1.UserResult
	(*ArticleResult)(nil),  // 3: search.v1.ArticleResult
	(*User)(nil),           // 4: search.v1.User
	(*Article)(nil),        // 5: search.v1.Article
}
var file_search_v1_search_proto_depIdxs = []int32{
	2, // 0: search.v1.SearchResponse.user:type_name -> search.v1.UserResult
	3, // 1: search.v1.SearchResponse.article:type_name -> search.v1.ArticleResult
	4, // 2: search.v1.UserResult.users:type_name -> search.v1.User
	5, // 3: search.v1.ArticleResult.articles:type_name -> search.v1.Article
	0, // 4: search.v1.SearchService.Search:input_type -> search.v1.SearchRequest
	1, // 5: search.v1.SearchService.Search:output_type -> search.v1.SearchResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_search_v1_search_proto_init() }
func file_search_v1_search_proto_init() {
	if File_search_v1_search_proto != nil {
		return
	}
	file_search_v1_sync_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_search_v1_search_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_search_v1_search_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_search_v1_search_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_search_v1_search_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ArticleResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_search_v1_search_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_search_v1_search_proto_goTypes,
		DependencyIndexes: file_search_v1_search_proto_depIdxs,
		MessageInfos:      file_search_v1_search_proto_msgTypes,
	}.Build()
	File_search_v1_search_proto = out.File
	file_search_v1_search_proto_rawDesc = nil
	file_search_v1_search_proto_goTypes = nil
	file_search_v1_search_proto_depIdxs = nil
}
