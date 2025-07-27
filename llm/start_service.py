#!/usr/bin/env python3
"""Simple startup script for the ML service."""

import os
import sys
import subprocess

def check_requirements():
    """Check if required packages are installed."""
    try:
        import fastapi
        import uvicorn
        import pydantic
        print("✅ Core packages available")
        return True
    except ImportError as e:
        print(f"❌ Missing required packages: {e}")
        print("Install with: pip install -r requirements.txt")
        return False

def check_env_file():
    """Check if .env file exists."""
    if os.path.exists('.env'):
        print("✅ .env file found")
        return True
    else:
        print("❌ .env file not found")
        print("Copy .env.example to .env and configure your API keys")
        return False

def main():
    """Start the ML service."""
    print("🚀 Starting RecMind ML Service...")
    
    if not check_env_file():
        return 1
    
    if not check_requirements():
        print("\nTrying to install requirements...")
        try:
            subprocess.check_call([sys.executable, "-m", "pip", "install", "-r", "requirements.txt"])
            print("✅ Requirements installed successfully")
        except subprocess.CalledProcessError:
            print("❌ Failed to install requirements")
            return 1
    
    print("\n🌟 Starting FastAPI server...")
    print("📖 API docs will be available at: http://localhost:8000/docs")
    print("🏥 Health check at: http://localhost:8000/health")
    print("\nPress Ctrl+C to stop the server\n")
    
    # Start the server
    try:
        subprocess.run([
            sys.executable, "-m", "uvicorn", 
            "app.main:app", 
            "--host", "0.0.0.0", 
            "--port", "8000", 
            "--reload"
        ])
    except KeyboardInterrupt:
        print("\n👋 Service stopped")
        return 0

if __name__ == "__main__":
    sys.exit(main())