import { useState, useEffect } from 'react';
import { getPosts, createPost } from '../services/api';

export default function Posts({ token, user }) {
    const [posts, setPosts] = useState([]);
    const [title, setTitle] = useState('');
    const [content, setContent] = useState('');

    useEffect(() => {
        const fetchPosts = async () => {
            try {
                const response = await getPosts();
                setPosts(response.data);
            } catch (err) {
                console.error('Failed to fetch posts', err);
            }
        };
        fetchPosts();
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await createPost({ title, content }, token);
            const response = await getPosts();
            setPosts(response.data);
            setTitle('');
            setContent('');
        } catch (err) {
            console.error('Failed to create post', err);
        }
    };

    return (
        <div>
            <h2>Posts</h2>
            {token && (
                <form onSubmit={handleSubmit}>
                    <input
                        type="text"
                        placeholder="Title"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        required
                    />
                    <textarea
                        placeholder="Content"
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        required
                    />
                    <button type="submit">Create Post</button>
                </form>
            )}
            <div>
                {posts.map(post => (
                    <div key={post.id} style={{ margin: '20px 0', padding: '10px', border: '1px solid #eee' }}>
                        <h3>{post.title}</h3>
                        <p>{post.content}</p>
                        <small>By: {post.author || 'Unknown'}</small>
                    </div>
                ))}
            </div>
        </div>
    );
}