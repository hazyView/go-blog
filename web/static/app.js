// Blog API Frontend Application
class BlogApp {
    constructor() {
        this.apiBase = '/api';
        this.currentUser = null;
        this.currentPost = null;
        this.posts = [];
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadPosts();
        this.showUserModal();
    }

    // Event Bindings
    bindEvents() {
        // Header buttons
        document.getElementById('newPostBtn').addEventListener('click', () => this.createNewPost());
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());

        // Editor buttons
        document.getElementById('saveBtn').addEventListener('click', () => this.savePost());
        document.getElementById('deleteBtn').addEventListener('click', () => this.deletePost());

        // Modal events
        document.getElementById('closeModal').addEventListener('click', () => this.hideUserModal());
        document.getElementById('userForm').addEventListener('submit', (e) => this.handleUserSubmit(e));
        document.getElementById('toggleMode').addEventListener('click', () => this.toggleUserMode());

        // Editor events
        document.getElementById('postTitle').addEventListener('input', () => this.updateWordCount());
        document.getElementById('postContent').addEventListener('input', () => this.updateWordCount());

        // Auto-save functionality
        setInterval(() => this.autoSave(), 30000); // Auto-save every 30 seconds

        // Click outside modal to close
        document.getElementById('userModal').addEventListener('click', (e) => {
            if (e.target.id === 'userModal') {
                this.hideUserModal();
            }
        });
    }

    // API Methods
    async apiCall(endpoint, method = 'GET', data = null) {
        this.showLoading();
        
        try {
            const config = {
                method,
                headers: {
                    'Content-Type': 'application/json',
                }
            };

            if (data) {
                config.body = JSON.stringify(data);
            }

            const response = await fetch(`${this.apiBase}${endpoint}`, config);
            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.message || result.error || 'API request failed');
            }

            return result;
        } catch (error) {
            this.showToast(error.message, 'error');
            throw error;
        } finally {
            this.hideLoading();
        }
    }

    // User Management
    async createUser(userData) {
        return await this.apiCall('/users', 'POST', userData);
    }

    async loginUser(username, password) {
        // Note: This is a simplified login. In production, you'd implement proper authentication
        const users = await this.apiCall('/users');
        const user = users.find(u => u.username === username);
        if (user) {
            this.currentUser = user;
            return user;
        }
        throw new Error('User not found');
    }

    // Post Management
    async loadPosts() {
        try {
            this.posts = await this.apiCall('/posts');
            this.renderPostsList();
        } catch (error) {
            this.showToast('Failed to load posts', 'error');
        }
    }

    async createPost(postData) {
        if (!this.currentUser) {
            this.showToast('Please log in first', 'error');
            return;
        }

        postData.user_id = this.currentUser.id;
        const post = await this.apiCall('/posts', 'POST', postData);
        this.posts.unshift(post);
        this.renderPostsList();
        this.showToast('Post created successfully!', 'success');
        return post;
    }

    async updatePost(id, postData) {
        const post = await this.apiCall(`/posts/${id}`, 'PUT', postData);
        const index = this.posts.findIndex(p => p.id === id);
        if (index !== -1) {
            this.posts[index] = post;
            this.renderPostsList();
        }
        this.showToast('Post updated successfully!', 'success');
        return post;
    }

    async deleteCurrentPost() {
        if (!this.currentPost) return;

        await this.apiCall(`/posts/${this.currentPost.id}`, 'DELETE');
        this.posts = this.posts.filter(p => p.id !== this.currentPost.id);
        this.renderPostsList();
        this.clearEditor();
        this.showToast('Post deleted successfully!', 'success');
    }

    // UI Methods
    renderPostsList() {
        const container = document.getElementById('postsList');
        
        if (this.posts.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-file-alt" style="font-size: 2rem; color: var(--text-muted); margin-bottom: 1rem;"></i>
                    <p style="color: var(--text-muted); text-align: center;">No posts yet. Create your first post!</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.posts.map(post => `
            <div class="post-item ${this.currentPost?.id === post.id ? 'active' : ''}" 
                 data-post-id="${post.id}">
                <i class="fas fa-star star-icon"></i>
                <div class="post-title">${this.escapeHtml(post.title)}</div>
                <div class="post-meta">
                    <span>${this.formatDate(post.created_at)}</span>
                </div>
            </div>
        `).join('');

        // Add click handlers
        container.querySelectorAll('.post-item').forEach(item => {
            item.addEventListener('click', () => {
                const postId = parseInt(item.dataset.postId);
                this.loadPost(postId);
            });
        });
    }

    loadPost(postId) {
        const post = this.posts.find(p => p.id === postId);
        if (!post) return;

        this.currentPost = post;
        document.getElementById('currentPostId').value = post.id;
        document.getElementById('postTitle').value = post.title;
        document.getElementById('postContent').value = post.content;
        
        this.updateWordCount();
        this.updateLastSaved(post.created_at);
        this.renderPostsList(); // Re-render to update active state
    }

    clearEditor() {
        this.currentPost = null;
        document.getElementById('currentPostId').value = '';
        document.getElementById('postTitle').value = '';
        document.getElementById('postContent').value = '';
        document.getElementById('lastSaved').textContent = 'Last saved: Never';
        this.updateWordCount();
        this.renderPostsList();
    }

    createNewPost() {
        if (!this.currentUser) {
            this.showUserModal();
            this.showToast('Please log in to create posts', 'info');
            return;
        }
        this.clearEditor();
        document.getElementById('postTitle').focus();
    }

    async savePost() {
        if (!this.currentUser) {
            this.showUserModal();
            return;
        }

        const title = document.getElementById('postTitle').value.trim();
        const content = document.getElementById('postContent').value.trim();

        if (!title || !content) {
            this.showToast('Please enter both title and content', 'error');
            return;
        }

        try {
            const postData = { title, content };

            if (this.currentPost) {
                // Update existing post
                const updatedPost = await this.updatePost(this.currentPost.id, postData);
                this.currentPost = updatedPost;
            } else {
                // Create new post
                const newPost = await this.createPost(postData);
                this.currentPost = newPost;
                document.getElementById('currentPostId').value = newPost.id;
            }

            this.updateLastSaved(new Date().toISOString());
        } catch (error) {
            // Error already handled in apiCall
        }
    }

    async deletePost() {
        if (!this.currentPost) {
            this.showToast('No post selected to delete', 'error');
            return;
        }

        if (confirm('Are you sure you want to delete this post? This action cannot be undone.')) {
            try {
                await this.deleteCurrentPost();
            } catch (error) {
                // Error already handled in apiCall
            }
        }
    }

    async autoSave() {
        if (!this.currentPost || !this.currentUser) return;

        const title = document.getElementById('postTitle').value.trim();
        const content = document.getElementById('postContent').value.trim();

        if (title && content && (title !== this.currentPost.title || content !== this.currentPost.content)) {
            try {
                const postData = { title, content };
                const updatedPost = await this.updatePost(this.currentPost.id, postData);
                this.currentPost = updatedPost;
                this.updateLastSaved(new Date().toISOString());
                this.showToast('Auto-saved', 'info');
            } catch (error) {
                // Fail silently for auto-save
            }
        }
    }

    // User Modal Methods
    showUserModal() {
        document.getElementById('userModal').classList.add('active');
    }

    hideUserModal() {
        document.getElementById('userModal').classList.remove('active');
        document.getElementById('userForm').reset();
    }

    toggleUserMode() {
        const modal = document.getElementById('userModal');
        const title = document.getElementById('modalTitle');
        const submitBtn = document.getElementById('submitBtn');
        const toggleBtn = document.getElementById('toggleMode');
        const emailGroup = document.querySelector('#email').closest('.form-group');

        if (title.textContent === 'Sign In') {
            title.textContent = 'Create Account';
            submitBtn.textContent = 'Create Account';
            toggleBtn.textContent = 'Sign In';
            emailGroup.style.display = 'block';
        } else {
            title.textContent = 'Sign In';
            submitBtn.textContent = 'Sign In';
            toggleBtn.textContent = 'Create Account';
            emailGroup.style.display = 'none';
        }
    }

    async handleUserSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const username = formData.get('username');
        const email = formData.get('email');
        const password = formData.get('password');

        const isSignUp = document.getElementById('modalTitle').textContent === 'Create Account';

        try {
            if (isSignUp) {
                await this.createUser({ username, email, password });
                this.showToast('Account created successfully! Please sign in.', 'success');
                this.toggleUserMode();
            } else {
                await this.loginUser(username, password);
                this.hideUserModal();
                this.showToast(`Welcome back, ${this.currentUser.username}!`, 'success');
                this.loadPosts();
            }
        } catch (error) {
            // Error already handled in apiCall
        }
    }

    logout() {
        this.currentUser = null;
        this.clearEditor();
        this.posts = [];
        this.renderPostsList();
        this.showUserModal();
        this.showToast('Logged out successfully', 'info');
    }

    // Utility Methods
    updateWordCount() {
        const content = document.getElementById('postContent').value;
        const wordCount = content.trim() ? content.trim().split(/\s+/).length : 0;
        document.getElementById('wordCount').textContent = `${wordCount} words`;
    }

    updateLastSaved(timestamp) {
        const date = new Date(timestamp);
        const formatted = date.toLocaleString();
        document.getElementById('lastSaved').textContent = `Last saved: ${formatted}`;
    }

    formatDate(timestamp) {
        const date = new Date(timestamp);
        const now = new Date();
        const diffMs = now - date;
        const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

        if (diffDays === 0) {
            return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        } else if (diffDays === 1) {
            return 'Yesterday';
        } else if (diffDays < 7) {
            return `${diffDays} days ago`;
        } else {
            return date.toLocaleDateString();
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    showLoading() {
        document.getElementById('loading').classList.add('active');
    }

    hideLoading() {
        document.getElementById('loading').classList.remove('active');
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;

        container.appendChild(toast);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            toast.style.animation = 'slideIn 0.3s ease reverse';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }, 5000);
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new BlogApp();
});
