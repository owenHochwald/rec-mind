#!/usr/bin/env python3
"""Test script for the ML service endpoints."""

import asyncio
import json
import uuid
from datetime import datetime

import httpx


BASE_URL = "http://localhost:8000"


async def test_health_endpoints():
    """Test health check endpoints."""
    print("🏥 Testing health endpoints...")
    
    async with httpx.AsyncClient() as client:
        # Basic health check
        response = await client.get(f"{BASE_URL}/health")
        print(f"Basic health: {response.status_code} - {response.json()}")
        
        # Detailed health check
        response = await client.get(f"{BASE_URL}/health/detail")
        print(f"Detailed health: {response.status_code}")
        if response.status_code == 200:
            health_data = response.json()
            print(f"  Service status: {health_data['service']['status']}")
            print(f"  Uptime: {health_data['uptime']:.2f}s")
            for dep in health_data['dependencies']:
                print(f"  {dep['name']}: {dep['status']}")


async def test_embedding_generation():
    """Test embedding generation."""
    print("\n🧠 Testing embedding generation...")
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        # Single embedding
        embedding_request = {
            "article_id": str(uuid.uuid4()),
            "text": "This is a test article about artificial intelligence and machine learning technologies."
        }
        
        response = await client.post(
            f"{BASE_URL}/embeddings/generate",
            json=embedding_request
        )
        
        if response.status_code == 200:
            result = response.json()
            print(f"Single embedding generated: {result['dimensions']} dimensions, {result['tokens_used']} tokens")
        else:
            print(f"Single embedding failed: {response.status_code} - {response.text}")


async def test_batch_embeddings():
    """Test batch embedding generation."""
    print("\n📚 Testing batch embedding generation...")
    
    async with httpx.AsyncClient(timeout=60.0) as client:
        # Batch embeddings
        batch_request = {
            "items": [
                {
                    "article_id": str(uuid.uuid4()),
                    "text": "Technology news about artificial intelligence developments."
                },
                {
                    "article_id": str(uuid.uuid4()),
                    "text": "Sports news covering the latest football match results."
                },
                {
                    "article_id": str(uuid.uuid4()),
                    "text": "Business news about market trends and financial analysis."
                }
            ]
        }
        
        response = await client.post(
            f"{BASE_URL}/embeddings/batch",
            json=batch_request
        )
        
        if response.status_code == 200:
            result = response.json()
            print(f"Batch embeddings generated: {len(result['results'])} items, {result['total_tokens']} tokens")
            print(f"Processing time: {result['processing_time']:.2f}s")
        else:
            print(f"Batch embeddings failed: {response.status_code} - {response.text}")


async def test_pinecone_operations():
    """Test Pinecone upload and search operations."""
    print("\n🔍 Testing Pinecone operations...")
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        # First generate an embedding to upload
        article_id = str(uuid.uuid4())
        embedding_request = {
            "article_id": article_id,
            "text": "Technology article about machine learning and AI advancements in 2025."
        }
        
        embed_response = await client.post(
            f"{BASE_URL}/embeddings/generate",
            json=embedding_request
        )
        
        if embed_response.status_code == 200:
            embedding_data = embed_response.json()
            
            # Upload to Pinecone
            upload_request = {
                "article_id": article_id,
                "embeddings": embedding_data["embeddings"],
                "metadata": {
                    "title": "Test AI Article",
                    "category": "Technology",
                    "created_at": datetime.utcnow().isoformat()
                }
            }
            
            upload_response = await client.post(
                f"{BASE_URL}/embeddings/upload",
                json=upload_request
            )
            
            if upload_response.status_code == 200:
                print(f"Vector uploaded successfully: {upload_response.json()}")
                
                # Wait a moment for indexing
                await asyncio.sleep(2)
                
                # Test similarity search
                search_request = {
                    "article_id": article_id,
                    "top_k": 3,
                    "score_threshold": 0.5
                }
                
                search_response = await client.post(
                    f"{BASE_URL}/search/similar",
                    json=search_request
                )
                
                if search_response.status_code == 200:
                    search_result = search_response.json()
                    print(f"Similarity search completed: {search_result['total_found']} results found")
                else:
                    print(f"Similarity search failed: {search_response.status_code} - {search_response.text}")
            else:
                print(f"Vector upload failed: {upload_response.status_code} - {upload_response.text}")
        else:
            print(f"Embedding generation for Pinecone test failed: {embed_response.status_code}")


async def test_index_stats():
    """Test index statistics endpoint."""
    print("\n📊 Testing index statistics...")
    
    async with httpx.AsyncClient() as client:
        response = await client.get(f"{BASE_URL}/index/stats")
        
        if response.status_code == 200:
            stats = response.json()
            print(f"Index stats: {stats['total_vectors']} vectors, dimension: {stats['dimension']}")
        else:
            print(f"Index stats failed: {response.status_code} - {response.text}")


async def main():
    """Run all tests."""
    print("🧪 Starting ML service tests...\n")
    
    try:
        await test_health_endpoints()
        await test_embedding_generation()
        await test_batch_embeddings()
        await test_pinecone_operations()
        await test_index_stats()
        
        print("\n✅ All tests completed!")
        
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")


if __name__ == "__main__":
    asyncio.run(main())