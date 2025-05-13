import axios from 'axios';

const API_BASE = 'http://localhost:8080/api/v1';
const FORUM_API = 'http://localhost:8081/api';

export const registerUser = (userData) => {
    return axios.post(`${API_BASE}/register`, userData);
};

export const loginUser = (credentials) => {
    return axios.post(`${API_BASE}/login`, credentials);
};

export const getPosts = () => {
    return axios.get(`${FORUM_API}/posts`);
};

export const createPost = (postData, token) => {
    return axios.post(`${FORUM_API}/posts`, postData, {
        headers: { Authorization: `Bearer ${token}` }
    });
};