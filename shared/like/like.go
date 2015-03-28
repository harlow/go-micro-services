package like

const ServiceID = "service.like"

type Args struct {
  PostID    int
  ServiceID      string
  TraceID string
  UserID    int
}

type Reply struct {
  ServiceID      string
  TraceID string
  Like Like
}

type Like struct {
  Count  int32
  PostID int
}
