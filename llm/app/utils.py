"""Utility functions to reduce code duplication."""

import functools
from typing import Callable, Any, Dict
from uuid import UUID

import structlog
from fastapi import HTTPException, Request


logger = structlog.get_logger(__name__)


def endpoint_error_handler(operation_name: str):
    """Decorator to handle common endpoint error patterns."""
    def decorator(func: Callable) -> Callable:
        @functools.wraps(func)
        async def wrapper(*args, **kwargs):
            try:
                # Extract article_id or other identifiers from kwargs for logging
                identifier = None
                for key, value in kwargs.items():
                    if hasattr(value, 'article_id'):
                        identifier = str(value.article_id)
                        break
                    elif isinstance(value, UUID):
                        identifier = str(value)
                        break
                
                if identifier:
                    logger.info(f"Processing {operation_name}", article_id=identifier)
                else:
                    logger.info(f"Processing {operation_name}")
                
                result = await func(*args, **kwargs)
                
                if identifier:
                    logger.info(f"{operation_name} completed successfully", article_id=identifier)
                else:
                    logger.info(f"{operation_name} completed successfully")
                
                return result
                
            except Exception as e:
                error_msg = f"{operation_name} failed: {str(e)}"
                if identifier:
                    logger.error(f"Failed to {operation_name.lower()}", article_id=identifier, error=str(e))
                else:
                    logger.error(f"Failed to {operation_name.lower()}", error=str(e))
                raise HTTPException(status_code=500, detail=error_msg)
        
        return wrapper
    return decorator


async def create_health_check_response(dependencies: list) -> Dict[str, Any]:
    """Create standardized health check response."""
    service_status = "healthy" if all(dep.status == "healthy" for dep in dependencies) else "unhealthy"
    
    return {
        "service_status": service_status,
        "dependencies": dependencies
    }


def get_request_identifier(request: Any) -> str:
    """Extract identifier from request for logging."""
    if hasattr(request, 'article_id'):
        return str(request.article_id)
    elif hasattr(request, 'id'):
        return str(request.id)
    return "unknown"