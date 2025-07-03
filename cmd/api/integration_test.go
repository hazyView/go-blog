package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"blog-api/internal/config"
	"blog-api/internal/database"
	"blog-api/internal/handlers"
	"blog-api/internal/models"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	server *httptest.Server
	db     *database.DB
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Configure logging for tests
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	// Load test configuration
	cfg := &config.Config{
		DatabaseHost: getEnv("TEST_DB_HOST", "localhost"),
		DatabasePort: getEnv("TEST_DB_PORT", "5432"),
		DatabaseUser: getEnv("TEST_DB_USER", "postgres"),
		DatabasePass: getEnv("TEST_DB_PASS", "password"),
		DatabaseName: getEnv("TEST_DB_NAME", "blog_api_test"),
	}

	// Initialize test database
	var err error
	suite.db, err = database.New(cfg)
	require.NoError(suite.T(), err)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(suite.db)
	postHandler := handlers.NewPostHandler(suite.db)
	healthHandler := handlers.NewHealthHandler(suite.db)

	// Setup test router
	router := setupRouter(userHandler, postHandler, healthHandler)

	// Create test server
	suite.server = httptest.NewServer(router)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up database before each test
	suite.cleanDatabase()
}

func (suite *IntegrationTestSuite) TestHealthEndpoint() {
	resp, err := http.Get(suite.server.URL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

func (suite *IntegrationTestSuite) TestUserCRUDOperations() {
	// Test Create User
	userReq := models.UserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Create user
	createdUser := suite.createUser(userReq)
	assert.Equal(suite.T(), userReq.Username, createdUser.Username)
	assert.Equal(suite.T(), userReq.Email, createdUser.Email)
	assert.NotZero(suite.T(), createdUser.ID)

	// Test Get User
	user := suite.getUser(createdUser.ID)
	assert.Equal(suite.T(), createdUser.ID, user.ID)
	assert.Equal(suite.T(), createdUser.Username, user.Username)

	// Test Update User
	updateReq := models.UserRequest{
		Username: "updateduser",
		Email:    "updated@example.com",
	}
	updatedUser := suite.updateUser(createdUser.ID, updateReq)
	assert.Equal(suite.T(), updateReq.Username, updatedUser.Username)
	assert.Equal(suite.T(), updateReq.Email, updatedUser.Email)

	// Test Get All Users
	users := suite.getAllUsers()
	assert.GreaterOrEqual(suite.T(), len(users), 1)

	// Test Delete User
	suite.deleteUser(createdUser.ID)
	
	// Verify user is deleted
	resp, err := http.Get(fmt.Sprintf("%s/users/%d", suite.server.URL, createdUser.ID))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

func (suite *IntegrationTestSuite) TestPostCRUDOperations() {
	// First create a user for the posts
	userReq := models.UserRequest{
		Username: "postauthor",
		Email:    "author@example.com",
		Password: "password123",
	}
	user := suite.createUser(userReq)

	// Test Create Post
	postReq := models.PostRequest{
		Title:   "Test Post",
		Content: "This is a test post content",
		UserID:  user.ID,
	}

	// Create post
	createdPost := suite.createPost(postReq)
	assert.Equal(suite.T(), postReq.Title, createdPost.Title)
	assert.Equal(suite.T(), postReq.Content, createdPost.Content)
	assert.Equal(suite.T(), postReq.UserID, createdPost.UserID)
	assert.NotZero(suite.T(), createdPost.ID)

	// Test Get Post
	post := suite.getPost(createdPost.ID)
	assert.Equal(suite.T(), createdPost.ID, post.ID)
	assert.Equal(suite.T(), createdPost.Title, post.Title)
	assert.Equal(suite.T(), user.Username, post.Username)

	// Test Update Post
	updateReq := models.PostRequest{
		Title:   "Updated Post",
		Content: "Updated content",
	}
	updatedPost := suite.updatePost(createdPost.ID, updateReq)
	assert.Equal(suite.T(), updateReq.Title, updatedPost.Title)
	assert.Equal(suite.T(), updateReq.Content, updatedPost.Content)

	// Test Get All Posts
	posts := suite.getAllPosts()
	assert.GreaterOrEqual(suite.T(), len(posts), 1)

	// Test Delete Post
	suite.deletePost(createdPost.ID)
	
	// Verify post is deleted
	resp, err := http.Get(fmt.Sprintf("%s/posts/%d", suite.server.URL, createdPost.ID))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

func (suite *IntegrationTestSuite) TestValidationErrors() {
	// Test invalid user creation
	invalidUser := models.UserRequest{
		Username: "", // Invalid: empty username
		Email:    "invalid-email", // Invalid: bad email format
		Password: "123", // Invalid: too short
	}

	userJSON, _ := json.Marshal(invalidUser)
	resp, err := http.Post(suite.server.URL+"/users", "application/json", bytes.NewBuffer(userJSON))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test invalid post creation
	invalidPost := models.PostRequest{
		Title:   "", // Invalid: empty title
		Content: "", // Invalid: empty content
		UserID:  0,  // Invalid: zero user ID
	}

	postJSON, _ := json.Marshal(invalidPost)
	resp, err = http.Post(suite.server.URL+"/posts", "application/json", bytes.NewBuffer(postJSON))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
}

// Helper methods for making HTTP requests

func (suite *IntegrationTestSuite) createUser(req models.UserRequest) models.User {
	userJSON, _ := json.Marshal(req)
	resp, err := http.Post(suite.server.URL+"/users", "application/json", bytes.NewBuffer(userJSON))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var user models.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(suite.T(), err)
	return user
}

func (suite *IntegrationTestSuite) getUser(id int) models.User {
	resp, err := http.Get(fmt.Sprintf("%s/users/%d", suite.server.URL, id))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var user models.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(suite.T(), err)
	return user
}

func (suite *IntegrationTestSuite) updateUser(id int, req models.UserRequest) models.User {
	userJSON, _ := json.Marshal(req)
	client := &http.Client{}
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("%s/users/%d", suite.server.URL, id), bytes.NewBuffer(userJSON))
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(httpReq)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var user models.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(suite.T(), err)
	return user
}

func (suite *IntegrationTestSuite) deleteUser(id int) {
	client := &http.Client{}
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/users/%d", suite.server.URL, id), nil)
	
	resp, err := client.Do(httpReq)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *IntegrationTestSuite) getAllUsers() []models.User {
	resp, err := http.Get(suite.server.URL + "/users")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var users []models.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	require.NoError(suite.T(), err)
	return users
}

func (suite *IntegrationTestSuite) createPost(req models.PostRequest) models.Post {
	postJSON, _ := json.Marshal(req)
	resp, err := http.Post(suite.server.URL+"/posts", "application/json", bytes.NewBuffer(postJSON))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var post models.Post
	err = json.NewDecoder(resp.Body).Decode(&post)
	require.NoError(suite.T(), err)
	return post
}

func (suite *IntegrationTestSuite) getPost(id int) models.Post {
	resp, err := http.Get(fmt.Sprintf("%s/posts/%d", suite.server.URL, id))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var post models.Post
	err = json.NewDecoder(resp.Body).Decode(&post)
	require.NoError(suite.T(), err)
	return post
}

func (suite *IntegrationTestSuite) updatePost(id int, req models.PostRequest) models.Post {
	postJSON, _ := json.Marshal(req)
	client := &http.Client{}
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("%s/posts/%d", suite.server.URL, id), bytes.NewBuffer(postJSON))
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(httpReq)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var post models.Post
	err = json.NewDecoder(resp.Body).Decode(&post)
	require.NoError(suite.T(), err)
	return post
}

func (suite *IntegrationTestSuite) deletePost(id int) {
	client := &http.Client{}
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/posts/%d", suite.server.URL, id), nil)
	
	resp, err := client.Do(httpReq)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *IntegrationTestSuite) getAllPosts() []models.Post {
	resp, err := http.Get(suite.server.URL + "/posts")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	require.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var posts []models.Post
	err = json.NewDecoder(resp.Body).Decode(&posts)
	require.NoError(suite.T(), err)
	return posts
}

func (suite *IntegrationTestSuite) cleanDatabase() {
	// Clean up posts first (due to foreign key constraint)
	suite.db.Exec("DELETE FROM posts")
	suite.db.Exec("DELETE FROM users")
	
	// Reset sequences
	suite.db.Exec("ALTER SEQUENCE posts_id_seq RESTART WITH 1")
	suite.db.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	suite.Run(t, new(IntegrationTestSuite))
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}
