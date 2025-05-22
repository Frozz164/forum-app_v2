import { useState, useEffect } from 'react';
import { forumAPI } from '../services/api';

export default function Posts({ token, user }) {
    const [posts, setPosts] = useState([]);
    const [title, setTitle] = useState('');
    const [content, setContent] = useState('');
    const [error, setError] = useState('');

    useEffect(() => {
        fetchPosts();
    }, []);

    const fetchPosts = async () => {
        try {
            const response = await forumAPI.getPosts();
            setPosts(response.data || []);
        } catch (err) {
            console.error('Error fetching posts:', err);
            setError(err.message);
        }
    };

    const createPost = async () => {
        if (!title.trim() || !content.trim()) {
            setError('Title and content are required');
            return;
        }

        try {
            await forumAPI.createPost({ title, content });
            setTitle('');
            setContent('');
            setError('');
            await fetchPosts();
        } catch (err) {
            console.error('Error creating post:', err);
            setError(err.message);
        }
    };

    const deletePost = async (postId) => {
        try {
            await forumAPI.deletePost(postId);
            await fetchPosts();
        } catch (err) {
            console.error('Error deleting post:', err);
            setError(err.message);
        }
    };

    return (
        <div className="posts-container">
            {token && (
                <div className="post-form">
                    <h3>Create New Post</h3>
                    {error && <div className="error-message">{error}</div>}
                    <input
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="Post title"
                        required
                    />
                    <textarea
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        placeholder="Post content"
                        required
                    />
                    <button onClick={createPost}>Create Post</button>
                </div>
            )}

            <div className="posts-list">
                <h3>Recent Posts</h3>
                {posts.map(post => (
                    <div key={post.id} className="post">
                        <h4>{post.title}</h4>
                        <p>{post.content}</p>
                        <div className="post-footer">
                            <span>By: {post.author || 'Unknown'}</span>
                            {post.author_id === user?.id && (
                                <button
                                    className="delete-btn"
                                    onClick={() => deletePost(post.id)}
                                >
                                    Delete
                                </button>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}