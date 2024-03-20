package events

const TopicReadEvent = "article_read"

const TopicGroupID = "interactive"

const Biz = "article"

type ReadEvent struct {
	Aid      int64
	Uid      int64
	ReadTime int64
}
