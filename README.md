<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <h1 align="center">RecMind</h1>
  <p align="center">
    A distributed semantic news search system with RAG architecture
    <br />
    Built with Go microservices, Python ML pipeline, and vector databases
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
        <li><a href="#built-with">Built With</a></li>
        <li><a href="#architecture">Architecture</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#api-reference">API Reference</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->
## About The Project

RecMind is a production-ready distributed semantic search system designed to demonstrate modern backend architecture patterns. It processes news articles to build a knowledge base, generates vector embeddings, and provides intelligent search recommendations through a clean React dashboard interacting with a REST API with async processing and job polling.

### Why This Architecture?

- **Distributed Design**: Chose microservices over monolith for service isolation, independent scaling, and technology flexibility
- **Async Processing**: RabbitMQ-based job queues prevent blocking operations and ensure reliability
- **Vector Search**: OpenAI embeddings + Pinecone deliver semantic understanding beyond simple keyword matching, supporting complex queries
- **Performance First**: Redis caching and connection pooling achieve sub-500ms search responses

### Key Features

- **Semantic Search**: Query-based recommendations using vector similarity
- **Dual Processing Modes**: Async (with polling) and immediate responses  
- **Smart Chunking**: Optimized text splitting for better search granularity
- **Fault Tolerance**: Circuit breakers, retry logic, graceful degradation
- **Real-time Monitoring**: Health checks with dependency status tracking

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Architecture

![](public/architecture.svg)



**System Flows:**

**Article Upload & Processing:**
1. React Dashboard → Go API `/api/upload` → PostgreSQL (store article metadata)
2. Go API → RabbitMQ `article_processing` queue → Python ML Service consumes job
3. Python ML → OpenAI API (generate embeddings) → Pinecone (store vectors with metadata)
4. Success/error status → Redis cache → React Dashboard updates via polling

**Search Query Processing:**
1. React Dashboard search → Go API `/api/v1/search/recommendations` → Create job in Redis
2. Go API → RabbitMQ `query_search_jobs` queue → Query RAG Worker consumes
3. Query RAG Worker → Python ML Service `/search/query` → OpenAI (query embedding)
4. Python ML → Pinecone vector search → Return similar chunks with scores
5. Query RAG Worker → PostgreSQL (enrich with article metadata) → Aggregate by article
6. Query RAG Worker → Redis (cache results) → Update job status → React polls for results
7. React Dashboard displays ranked articles with similarity scores and request timings

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

**Frontend:**
* [![React][React.js]][React-url] - Interactive dashboard for article services

**Backend Services:**
* [![Go][Go.dev]][Go-url] - API server, workers, and distributed processing
* [![Python][Python.org]][Python-url] - ML service with FastAPI and embeddings
* [![PostgreSQL][PostgreSQL.org]][PostgreSQL-url] - Primary database 
* [![Redis][Redis.io]][Redis-url] - Caching and job status tracking
* [![RabbitMQ][RabbitMQ.com]][RabbitMQ-url] - Async message queue and job processing

**AI & Vector Search:**
* [![OpenAI][OpenAI.com]][OpenAI-url] - text-embedding-3-small model
* [![Pinecone][Pinecone.io]][Pinecone-url] - Vector database for semantic search

**Infrastructure:**
* [![Docker][Docker.com]][Docker-url] - Containerized microservices
* [![Gin][Gin-Gonic.com]][Gin-url] - High-performance Go web framework

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

* **Go 1.21+** for API server and workers (for manual setup)
* **Python 3.9+** for ML service (for manual setup)
* **Docker & Docker Compose** for containerized deployment
* **API Keys**: OpenAI and Pinecone accounts

### Quick Start with Docker (Recommended)

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

3. **Launch the complete stack**
   ```bash
   docker-compose up -d
   ```

4. **Access the application**
   - **React Dashboard**: http://localhost:3000
   - **API Documentation**: http://localhost:8080/swagger/index.html
   - **RabbitMQ Management**: http://localhost:15672 (myuser/secret)

5. **Check system health**
   ```bash
   docker-compose ps
   curl http://localhost:8080/health
   ```

### Manual Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/owenhochwald/rec-mind.git
   cd rec-mind
   ```

2. **Start infrastructure services**
   ```bash
   cd infra
   docker-compose up -d  # PostgreSQL, Redis, RabbitMQ
   ```

3. **Configure environment variables**
   ```bash
   # Go API (.env in /api directory)
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=password
   DB_NAME=rec_mind
   REDIS_URL=redis://localhost:6379
   RABBITMQ_URL=amqp://guest:guest@localhost:5672/
   PYTHON_ML_SERVICE_URL=http://localhost:8000

   # Python ML (.env in /llm directory)
   OPENAI_API_KEY=your_openai_api_key
   PINECONE_API_KEY=your_pinecone_api_key
   PINECONE_INDEX_NAME=your_index_name
   RABBITMQ_URL=amqp://guest:guest@localhost:5672/
   ```

4. **Install Python dependencies**
   ```bash
   cd llm
   pip3 install -r requirements.txt
   ```

5. **Install React frontend dependencies**
   ```bash
   cd frontend
   npm install
   ```

6. **Start the services**
   ```bash
   # Terminal 1: Python ML Service
   cd llm && python3 -m uvicorn app.main:app --reload

   # Terminal 2: Go API Server  
   cd api && go run server/main.go

   # Terminal 3: Query RAG Worker
   cd api && go run cmd/query_rag_worker/main.go

   # Terminal 4: React Frontend
   cd frontend && npm start
   ```

7. **Access the application**
   ```bash
   open http://localhost:3000                     # React Dashboard
   open http://localhost:8080/swagger/index.html  # API Documentation
   curl http://localhost:8080/health              # Health Check
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

### React Dashboard (Recommended)
The easiest way to interact with RecMind is through the React dashboard at `http://localhost:3000`:

- **Upload Articles**: Interactive form with real-time processing status
- **Search Articles**: Query input with live results and similarity scores  
- **Manage Articles**: Browse, filter, and delete articles with pagination
- **System Health**: Monitor all services and their status
- **Request Timing**: View API call durations and performance metrics

_Complete API documentation: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)_

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- API REFERENCE -->
## API Reference

### Core Endpoints
- `GET /health` - System health with dependency status
- `POST /api/upload` - Upload articles with ML processing  
- `POST /api/v1/search/recommendations` - Query-based search (async)
- `POST /api/v1/search/immediate` - Query-based search (sync)
- `GET /api/v1/search/jobs/:job_id` - Check search job status

### Article Management
- `GET /api/v1/articles` - List articles with pagination
- `GET /api/v1/articles/:id` - Get article details
- `DELETE /api/v1/articles/:id` - Delete article

### Development Commands
```bash
make dev      # Run Go server without building
make build    # Build binary  
make test     # Run tests
make deps     # Update dependencies
```

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