import React, { useState, useEffect } from 'react';
import { apiClient, Article, ApiRequestTiming, HealthStatus } from '../services/api';

const Dashboard: React.FC = () => {
  // TODO: Practice useState - Create state for:
  // const [articles, setArticles] = useState<Article[]>([]);
  // const [loading, setLoading] = useState<boolean>(false);
  // const [error, setError] = useState<string | null>(null);
  // const [lastRequestTiming, setLastRequestTiming] = useState<ApiRequestTiming | null>(null);

  const [activeSection, setActiveSection] = useState<string>('articles');

  // TODO: Practice useEffect - Load initial data
  // useEffect(() => {
  //   // Load articles on component mount
  //   loadArticles();
  // }, []);

  // TODO: Create loadArticles function that uses apiClient.getArticles()
  const loadArticles = async () => {
    console.log('TODO: Implement loadArticles using apiClient.getArticles()');
  };

  // TODO: Create health check function with useEffect polling
  const checkHealth = async () => {
    console.log('TODO: Implement health check using apiClient.getHealth()');
  };

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>RecMind Dashboard</h1>
        <div className="request-timing">
          {/* TODO: Display last request timing info here */}
          Last Request: -- ms
        </div>
      </div>

      <div className="dashboard-content">
        {/* Simple Navigation */}
        <nav className="dashboard-nav">
          <button 
            className={activeSection === 'articles' ? 'active' : ''}
            onClick={() => setActiveSection('articles')}
          >
            Articles
          </button>
          <button 
            className={activeSection === 'upload' ? 'active' : ''}
            onClick={() => setActiveSection('upload')}
          >
            Upload
          </button>
          <button 
            className={activeSection === 'search' ? 'active' : ''}
            onClick={() => setActiveSection('search')}
          >
            Search
          </button>
          <button 
            className={activeSection === 'health' ? 'active' : ''}
            onClick={() => setActiveSection('health')}
          >
            Health
          </button>
        </nav>

        <main className="dashboard-main">
          {activeSection === 'articles' && (
            <ArticlesSection />
          )}
          
          {activeSection === 'upload' && (
            <UploadSection />
          )}
          
          {activeSection === 'search' && (
            <SearchSection />
          )}
          
          {activeSection === 'health' && (
            <HealthSection />
          )}
        </main>
      </div>
    </div>
  );
};

// TODO: Implement these components with useState and useEffect
const ArticlesSection: React.FC = () => {
  // TODO: Practice useState for articles list, loading, selectedArticle
  // TODO: Practice useEffect to load articles on mount
  // TODO: Add "Load Chunks" button for each article

  return (
    <div className="articles-section">
      <h2>Articles</h2>
      <p>TODO: Display articles list with pagination</p>
      <p>TODO: Add "Load Chunks" button for each article</p>
      <p>TODO: Add delete functionality with confirmation</p>
      
      {/* Sample structure:
      <div className="articles-list">
        {articles.map(article => (
          <div key={article.id} className="article-item">
            <h3>{article.title}</h3>
            <p>{article.category} - {article.published_at}</p>
            <button onClick={() => loadChunks(article.id)}>Load Chunks</button>
            <button onClick={() => deleteArticle(article.id)}>Delete</button>
          </div>
        ))}
      </div>
      */}
    </div>
  );
};

const UploadSection: React.FC = () => {
  // TODO: Practice useState for form data (title, content, url, category, published_at)
  // TODO: Add state for processing mode (async/sync)
  // TODO: Add state for upload progress and results

  return (
    <div className="upload-section">
      <h2>Upload Article</h2>
      <p>TODO: Create form with validation</p>
      <p>TODO: Add async/sync processing toggle</p>
      <p>TODO: Show upload progress and results</p>
      
      {/* Sample form structure:
      <form onSubmit={handleSubmit}>
        <input type="text" placeholder="Title" />
        <textarea placeholder="Content"></textarea>
        <input type="url" placeholder="URL" />
        <input type="text" placeholder="Category" />
        <input type="datetime-local" />
        <label>
          <input type="checkbox" /> Sync Processing
        </label>
        <button type="submit">Upload</button>
      </form>
      */}
    </div>
  );
};

const SearchSection: React.FC = () => {
  // TODO: Practice useState for query, results, searchMode (async/immediate)
  // TODO: Practice useEffect for polling async search jobs
  // TODO: Add state for search parameters (top_k, score_threshold)

  return (
    <div className="search-section">
      <h2>Search Recommendations</h2>
      <p>TODO: Create search interface with parameters</p>
      <p>TODO: Add async/immediate search toggle</p>
      <p>TODO: Display results with similarity scores</p>
      <p>TODO: Add job polling for async searches</p>
      
      {/* Sample search structure:
      <div className="search-form">
        <input type="text" placeholder="Search query..." />
        <input type="number" placeholder="Top K" defaultValue="10" />
        <input type="number" step="0.1" placeholder="Score Threshold" defaultValue="0.7" />
        <label>
          <input type="radio" name="searchMode" value="immediate" /> Immediate
          <input type="radio" name="searchMode" value="async" /> Async
        </label>
        <button>Search</button>
      </div>
      */}
    </div>
  );
};

const HealthSection: React.FC = () => {
  // TODO: Practice useState for health status
  // TODO: Practice useEffect with setInterval for real-time updates
  // TODO: Add color-coded status indicators

  return (
    <div className="health-section">
      <h2>System Health</h2>
      <p>TODO: Display service status with color indicators</p>
      <p>TODO: Add real-time updates with useEffect + setInterval</p>
      
      {/* Sample health display:
      <div className="health-status">
        <div className="status-item">
          <span className="status-indicator green"></span>
          API: Online
        </div>
        <div className="status-item">
          <span className="status-indicator green"></span>
          Database: Connected
        </div>
        // ... more services
      </div>
      */}
    </div>
  );
};

export default Dashboard;