package database

import (
	"context"
	"testing"
	"time"

	"blog-api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserOperations tests all user CRUD operations
func TestUserOperations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()

	t.Run("CreateUser", func(t *testing.T) {
		req := &models.UserRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		user, err := db.CreateUser(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NotZero(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// First create a user
		req := &models.UserRequest{
			Username: "getuser",
			Email:    "get@example.com",
			Password: "password123",
		}
		createdUser, err := db.CreateUser(ctx, req)
		require.NoError(t, err)

		// Then get the user
		user, err := db.GetUserByID(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, createdUser.Username, user.Username)
		assert.Equal(t, createdUser.Email, user.Email)
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Create multiple users
		users := []models.UserRequest{
			{Username: "user1", Email: "user1@example.com", Password: "password123"},
			{Username: "user2", Email: "user2@example.com", Password: "password123"},
		}

		for _, userReq := range users {
			_, err := db.CreateUser(ctx, &userReq)
			require.NoError(t, err)
		}

		// Get all users
		allUsers, err := db.GetAllUsers(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allUsers), 2)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		// Create a user
		req := &models.UserRequest{
			Username: "updateuser",
			Email:    "update@example.com",
			Password: "password123",
		}
		createdUser, err := db.CreateUser(ctx, req)
		require.NoError(t, err)

		// Update the user
		updateReq := &models.UserRequest{
			Username: "updateduser",
			Email:    "updated@example.com",
		}
		updatedUser, err := db.UpdateUser(ctx, createdUser.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", updatedUser.Username)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		// Create a user
		req := &models.UserRequest{
			Username: "deleteuser",
			Email:    "delete@example.com",
			Password: "password123",
		}
		createdUser, err := db.CreateUser(ctx, req)
		require.NoError(t, err)

		// Delete the user
		err = db.DeleteUser(ctx, createdUser.ID)
		require.NoError(t, err)

		// Try to get the deleted user
		_, err = db.GetUserByID(ctx, createdUser.ID)
		assert.Error(t, err)
	})
}

// TestPostOperations tests all post CRUD operations
func TestPostOperations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()

	// First create a user for the posts
	userReq := &models.UserRequest{
		Username: "postauthor",
		Email:    "author@example.com",
		Password: "password123",
	}
	user, err := db.CreateUser(ctx, userReq)
	require.NoError(t, err)

	t.Run("CreatePost", func(t *testing.T) {
		req := &models.PostRequest{
			Title:   "Test Post",
			Content: "This is a test post content",
			UserID:  user.ID,
		}

		post, err := db.CreatePost(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "Test Post", post.Title)
		assert.Equal(t, "This is a test post content", post.Content)
		assert.Equal(t, user.ID, post.UserID)
		assert.NotZero(t, post.ID)
		assert.False(t, post.CreatedAt.IsZero())
	})

	t.Run("GetPostByID", func(t *testing.T) {
		// Create a post
		req := &models.PostRequest{
			Title:   "Get Post Test",
			Content: "Content for get test",
			UserID:  user.ID,
		}
		createdPost, err := db.CreatePost(ctx, req)
		require.NoError(t, err)

		// Get the post
		post, err := db.GetPostByID(ctx, createdPost.ID)
		require.NoError(t, err)
		assert.Equal(t, createdPost.ID, post.ID)
		assert.Equal(t, createdPost.Title, post.Title)
		assert.Equal(t, createdPost.Content, post.Content)
		assert.Equal(t, user.Username, post.Username)
	})

	t.Run("GetAllPosts", func(t *testing.T) {
		// Create multiple posts
		posts := []models.PostRequest{
			{Title: "Post 1", Content: "Content 1", UserID: user.ID},
			{Title: "Post 2", Content: "Content 2", UserID: user.ID},
		}

		for _, postReq := range posts {
			_, err := db.CreatePost(ctx, &postReq)
			require.NoError(t, err)
		}

		// Get all posts
		allPosts, err := db.GetAllPosts(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allPosts), 2)
	})

	t.Run("UpdatePost", func(t *testing.T) {
		// Create a post
		req := &models.PostRequest{
			Title:   "Update Post",
			Content: "Original content",
			UserID:  user.ID,
		}
		createdPost, err := db.CreatePost(ctx, req)
		require.NoError(t, err)

		// Update the post
		updateReq := &models.PostRequest{
			Title:   "Updated Post",
			Content: "Updated content",
		}
		updatedPost, err := db.UpdatePost(ctx, createdPost.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "Updated Post", updatedPost.Title)
		assert.Equal(t, "Updated content", updatedPost.Content)
	})

	t.Run("DeletePost", func(t *testing.T) {
		// Create a post
		req := &models.PostRequest{
			Title:   "Delete Post",
			Content: "Content to delete",
			UserID:  user.ID,
		}
		createdPost, err := db.CreatePost(ctx, req)
		require.NoError(t, err)

		// Delete the post
		err = db.DeletePost(ctx, createdPost.ID)
		require.NoError(t, err)

		// Try to get the deleted post
		_, err = db.GetPostByID(ctx, createdPost.ID)
		assert.Error(t, err)
	})
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *DB {
	// Note: This is a simplified setup for unit tests
	// In a real scenario, you would use a test database or an in-memory database
	// For now, this is a placeholder that assumes a test database is available
	
	// You might want to use environment variables or a separate test config
	// to connect to a test database
	t.Skip("Test database setup required - please configure test database connection")
	return nil
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, db *DB) {
	if db != nil {
		db.Close()
	}
}
