import axios from 'axios';

const API_BASE = 'http://localhost:8080/api/v1';
const FORUM_API = 'http://localhost:8081/api';

const apiClient = axios.create({
    baseURL: API_BASE,
    withCredentials: true,
    timeout: 10000
});

const forumClient = axios.create({
    baseURL: FORUM_API,
    timeout: 10000
});

// Интерсептор для добавления токена
forumClient.interceptors.request.use(config => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    } else {
        delete config.headers.Authorization;
    }
    return config;
});

// Интерсептор для обработки 401 ошибки
forumClient.interceptors.response.use(
    response => response,
    error => {
        if (error.response?.status === 401) {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.reload();
        }
        return Promise.reject(error);
    }
);

const handleError = (error) => {
    if (error.response) {
        console.error('API Error:', error.response.data);
        throw new Error(error.response.data.error || 'Request failed');
    } else {
        console.error('Network Error:', error.message);
        throw new Error('Network error. Please try again.');
    }
};

export const authAPI = {
    register: (userData) => apiClient.post('/register', userData).catch(handleError),
    login: (credentials) => apiClient.post('/login', credentials).catch(handleError)
};

export const forumAPI = {
    getPosts: () => forumClient.get('/posts').catch(handleError),
    createPost: (postData) => forumClient.post('/posts', postData).catch(handleError),
    deletePost: (postId) => forumClient.delete(`/posts/${postId}`).catch(handleError)
};

// Для совместимости
export const registerUser = authAPI.register;
export const loginUser = authAPI.login;
export const getPosts = forumAPI.getPosts;
export const createPost = forumAPI.createPost;
export const deletePost = forumAPI.deletePost;