<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <h1 align="center">RecMind</h1>
  <p align="center">
    A private, self-hosted knowledge base system for intelligent content discovery
    <br />
    Like NotebookLM, but for your personal research and reading collection
    <br />
    <br />

  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#the-problem">The Problem</a></li>
        <li><a href="#the-solution">The Solution</a></li>
        <li><a href="#built-with">Built With</a></li>
        <li><a href="#architecture">Architecture</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#quick-start">Quick Start</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#technical-highlights">Technical Highlights</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project

### The Problem

Ever had a vague memory of reading something perfect for your current project, but couldn't find it? You know it was about "that distributed systems thing with eventual consistency," but searching through bookmarks, notes, and browser history yields nothing useful.

**RecMind solves the "where did I see that article?" problem.**

### The Solution

RecMind is a **real-time knowledge base system** that lets you build a massive personal library of completely unrelated content: tech blogs, research papers, random posts, documentation, and query it semantically to surface exactly what you need.

Think of it as your private, self-hosted NotebookLM:
- **Semantic Understanding**: Ask "microservices communication patterns" and find relevant content even if those exact words weren't used
- **Completely Private**: Your data stays on your infrastructure
- **Real-time Search**: Sub-600ms responses with intelligent caching
- **Production Architecture**: Built with the same patterns used by major tech companies


<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Architecture

![](public/architecture.svg)

**Why This Distributed Architecture?**


- **Message Queues**: RabbitMQ prevents blocking operations and ensures reliability—critical for user experience
- **Microservices**: Independent scaling and technology choices (Go for performance, Python for ML)
- **Async Processing**: Background jobs keep the UI responsive while processing heavy ML workloads
- **Vector Search**: OpenAI embeddings + Pinecone deliver semantic understanding that keyword search can't match
- **Multi-layer Caching**: Redis + connection pooling achieve enterprise-level performance

**System Flows:**

**Content Processing:**
1. Upload article → Go API → PostgreSQL (metadata) → RabbitMQ job queue
2. Python ML service → OpenAI embeddings → Pinecone vector storage
3. Real-time status updates via Redis → React dashboard

**Intelligent Search:**
1. Semantic query → Go API → RabbitMQ → Query worker
2. Query worker → Python ML → OpenAI embedding → Pinecone similarity search
3. Results aggregation → PostgreSQL enrichment → Redis caching → UI updates

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

**Frontend:**
* [![React][React.js]][React-url] - Modern dashboard with real-time updates

**Backend Services:**
* [![Go][Go.dev]][Go-url] - High-performance API server and workers
* [![Python][Python.org]][Python-url] - ML service with LangChain and FastAPI
* [![PostgreSQL][PostgreSQL.org]][PostgreSQL-url] - Primary database with full-text search
* [![Redis][Redis.io]][Redis-url] - Caching and job status tracking
* [![RabbitMQ][RabbitMQ.com]][RabbitMQ-url] - Async message queue and job processing

**AI & Vector Search:**
* [![OpenAI][OpenAI.com]][OpenAI-url] - text-embedding-3-small for semantic understanding
* [![Pinecone][Pinecone.io]][Pinecone-url] - Vector database for similarity search

**Infrastructure:**
* [![Docker][Docker.com]][Docker-url] - Complete containerized stack (6 services)
* [![Gin][Gin-Gonic.com]][Gin-url] - Production-grade Go web framework

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

* **Docker & Docker Compose** (recommended)
* **API Keys**: OpenAI and Pinecone accounts
* **Go 1.21+** and **Python 3.9+** (for manual setup)

### Quick Start

1. **Clone and configure**
   ```bash
   git clone https://github.com/owenhochwald/rec-mind.git
   cd rec-mind
   ```

2. **Add your API keys to `.env`**
   ```bash
   OPENAI_API_KEY=your_openai_api_key_here
   PINECONE_API_KEY=your_pinecone_api_key_here
   PINECONE_INDEX_NAME=your_pinecone_index_name
   ```

3. **Launch the complete stack (6 services)**
   ```bash
   docker compose up -d
   ```

4. **Start exploring**
   - **React Dashboard**: http://localhost:3000
   - **API Documentation**: http://localhost:8080/swagger/index.html
   - **System Health**: http://localhost:8080/health

5. **Upload your first article and search!**

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

### Real-World Scenarios

**Research Repository**
Upload papers, blogs, and documentation. Query "distributed consensus algorithms" to surface everything relevant, even if the exact phrase wasn't used.

**Idea Connection**
Save random articles and posts. Later, search "event sourcing patterns" and rediscover that perfect Martin Fowler post you read months ago.

**Learning Enhancement**
Build a personal knowledge base as you learn. Query concepts to see how different sources explain the same ideas.

### Using the Dashboard

The React dashboard at `http://localhost:3000` provides:

- **Smart Upload**: Add articles with real-time processing status
- **Semantic Search**: Query your knowledge base with natural language
- **Content Management**: Browse, filter, and organize your articles
- **Performance Monitoring**: View response times and system health
- **Relevance Scoring**: See how closely results match your query

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Technical Highlights

### Distributed Systems Patterns

- **Event-Driven Architecture**: Async processing with message queues
- **Circuit Breakers**: Fault tolerance and graceful degradation
- **Horizontal Scaling**: Independent service scaling based on load
- **Observability**: Health checks and performance monitoring

### Performance Engineering

- **Sub-600ms Search**: Multi-layer caching and connection pooling
- **Async Processing**: Non-blocking ML operations
- **Batch Optimization**: Efficient vector operations and database queries
- **Resource Management**: Connection limits and timeout handling

### Production Readiness

- **Containerized Deployment**: Single-command infrastructure setup
- **Environment Configuration**: Proper secrets and config management
- **API Documentation**: Auto-generated Swagger/OpenAPI specs

<p align="right">(<a href="#readme-top">back to top</a>)</p>


### Development Commands
```bash
make dev      # Run Go server without building
make build    # Build binary  
make test     # Run tests
docker compose logs -f  # View all service logs
```

**Complete API Documentation**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
[contributors-shield]: https://img.shields.io/github/contributors/owenhochwald/rec-mind.svg?style=for-the-badge
[contributors-url]: https://github.com/owenhochwald/rec-mind/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/owenhochwald/rec-mind.svg?style=for-the-badge
[forks-url]: https://github.com/owenhochwald/rec-mind/network/members
[stars-shield]: https://img.shields.io/github/stars/owenhochwald/rec-mind.svg?style=for-the-badge
[stars-url]: https://github.com/owenhochwald/rec-mind/stargazers
[issues-shield]: https://img.shields.io/github/issues/owenhochwald/rec-mind.svg?style=for-the-badge
[issues-url]: https://github.com/owenhochwald/rec-mind/issues
[license-shield]: https://img.shields.io/github/license/owenhochwald/rec-mind.svg?style=for-the-badge
[license-url]: https://github.com/owenhochwald/rec-mind/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/owenhochwald

<!-- Technology Badges -->
[React.js]: https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB
[React-url]: https://reactjs.org/
[Go.dev]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://golang.org/
[Python.org]: https://img.shields.io/badge/Python-3776AB?style=for-the-badge&logo=python&logoColor=white
[Python-url]: https://python.org/
[PostgreSQL.org]: https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white
[PostgreSQL-url]: https://postgresql.org/
[Redis.io]: https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white
[Redis-url]: https://redis.io/
[RabbitMQ.com]: https://img.shields.io/badge/RabbitMQ-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white
[RabbitMQ-url]: https://rabbitmq.com/
[OpenAI.com]: https://img.shields.io/badge/OpenAI-412991?style=for-the-badge&logo=openai&logoColor=white
[OpenAI-url]: https://openai.com/
[Pinecone.io]: https://img.shields.io/badge/Pinecone-000000?style=for-the-badge&logo=pinecone&logoColor=white
[Pinecone-url]: https://pinecone.io/
[Docker.com]: https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white
[Docker-url]: https://docker.com/
[Gin-Gonic.com]: https://img.shields.io/badge/Gin-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Gin-url]: https://gin-gonic.com/