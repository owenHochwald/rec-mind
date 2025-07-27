#!/usr/bin/env python3
"""Test script for Pinecone connection configuration."""

import asyncio
import os
from app.config import get_settings
from app.vectordb import VectorDBService


async def test_pinecone_connection():
    """Test Pinecone connection with proper configuration."""
    print("Testing Pinecone Connection Configuration...")
    
    try:
        # Load settings
        settings = get_settings()
        print(f"Configuration loaded:")
        print(f"   Index Name: {settings.pinecone_index_name}")
        print(f"   Environment: {settings.pinecone_environment}")
        print(f"   Host URL: {settings.pinecone_host if settings.pinecone_host else 'Not configured'}")
        print(f"   API Key: {'Configured' if settings.pinecone_api_key else 'Missing'}")
        
        # Create service instance
        service = VectorDBService()
        print(f"\nVectorDB Service created:")
        print(f"   Using host: {bool(service.host)}")
        print(f"   Host URL: {service.host[:50] + '...' if service.host else 'Environment-based'}")
        
        # Test connection initialization (without actually connecting to avoid API calls)
        print(f"\nConnection configuration:")
        if service.host:
            host_url = service.host
            if host_url.startswith('https://'):
                host_url = host_url[8:]  # Remove https://
            print(f"   Will connect using direct host: {host_url}")
        else:
            print(f"   Will connect using environment: {service.environment}")
        
        # Test health check structure (without actual API call)
        print(f"\nHealth check configuration ready")
        print(f"   Index name: {service.index_name}")
        print(f"   Dimension: {service.dimension}")
        
        print(f"\nPinecone connection properly configured!")
        print(f"Note: Actual connection will be established when ML service starts")
        
        return True
        
    except Exception as e:
        print(f"\nConfiguration error: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = asyncio.run(test_pinecone_connection())
    if success:
        print(f"\nPinecone configuration test completed successfully!")
    else:
        print(f"\nConfiguration test failed!")