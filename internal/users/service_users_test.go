package users

// import (
// 	"context"
// 	"errors"
// 	"testing"
// )
// TODO: remake test file due to change in UserRegistration, logic tested with docker + curl via bash | works
// type mockRepo struct {
// 	emailExists      bool
// 	emailExistsErr   error
// 	createUserErr    error
// 	createUserCalled bool
// 	lastUser         *User
// }

// func (m *mockRepo) EmailExists(ctx context.Context, email string) (bool, error) {
// 	return m.emailExists, m.emailExistsErr
// }

// func (m *mockRepo) CreateUser(ctx context.Context, user *User) error {
// 	m.createUserCalled = true
// 	return m.createUserErr
// }

// // TESTS

// // Happy case, all works
// func TestUserRegistration_Ok(t *testing.T) {
// 	mock := &mockRepo{
// 		emailExists: false,
// 	}

// 	service := NewService(mock)

// 	data := &Registration{
// 		Username: "testWorks",
// 		Email:    "test@test.com",
// 		Password: "testpassword",
// 	}

// 	err := service.UserRegistration(context.Background(), data)
// 	if err != nil {
// 		t.Errorf("expected no error, got: %v", err)
// 	}

// 	if !mock.createUserCalled {
// 		t.Error("expected CreateUser to be called")
// 	}
// }

// // Case email exists, should not call createUser
// func TestUserRegistration_EmailExists(t *testing.T) {
// 	mock := &mockRepo{
// 		emailExists: true,
// 	}
// 	service := NewService(mock)

// 	data := &Registration{
// 		Username: "test1",
// 		Email:    "test@test.com",
// 		Password: "testpassword",
// 	}

// 	err := service.UserRegistration(context.Background(), data)

// 	if !errors.Is(err, ErrEmailAlreadyExists) {
// 		t.Errorf("expected ErrEmailAlreadyExists, got: %v", err)
// 	}

// 	if mock.createUserCalled {
// 		t.Error("expected CreateUser NOT to be called")
// 	}
// }

// func TestUserRegistration_EmailExistsFails(t *testing.T) {
// 	mock := &mockRepo{
// 		emailExistsErr: errors.New("db error"),
// 	}

// 	service := NewService(mock)

// 	data := &Registration{
// 		Username: "test2",
// 		Email:    "test@test.com",
// 		Password: "testpassword",
// 	}

// 	err := service.UserRegistration(context.Background(), data)

// 	if err == nil {
// 		t.Error("expected error, got nil")
// 	}
// }

// func TestUserRegistration_CreateUserFails(t *testing.T) {
// 	mock := &mockRepo{
// 		emailExists:   false,
// 		createUserErr: errors.New("insert error"),
// 	}

// 	service := NewService(mock)

// 	data := &Registration{
// 		Username: "test3",
// 		Email:    "test@test.com",
// 		Password: "testpassword",
// 	}

// 	err := service.UserRegistration(context.Background(), data)

// 	if err == nil {
// 		t.Error("expected error got nil")
// 	}

// 	if !mock.createUserCalled {
// 		t.Error("expected CreateUser to be called")
// 	}
// }
