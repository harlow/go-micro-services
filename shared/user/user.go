package user

const ServiceID = "service.user"

type Args struct {
  AuthToken string
  ServiceID string
  TraceID   string
}

type Reply struct {
  ServiceID string
  TraceID   string
  User      User
}

type User struct {
  Email     string
  FirstName string
  ID        int
  LastName  string
}
