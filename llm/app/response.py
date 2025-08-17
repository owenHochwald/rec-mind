"""Standardized response helpers for FastAPI."""

from typing import Any, Dict, Optional
from fastapi import HTTPException
from fastapi.responses import JSONResponse


class APIResponse:
    """Standardized API response helper."""
    
    @staticmethod
    def success(data: Any, message: Optional[str] = None) -> Dict[str, Any]:
        """Create a success response."""
        response = {
            "success": True,
            "data": data
        }
        if message:
            response["message"] = message
        return response
    
    @staticmethod
    def error(
        message: str,
        error_code: Optional[str] = None,
        details: Optional[Any] = None,
        status_code: int = 400
    ) -> HTTPException:
        """Create an error response."""
        error_data = {
            "success": False,
            "error": message
        }
        if error_code:
            error_data["error_code"] = error_code
        if details:
            error_data["details"] = details
        
        raise HTTPException(status_code=status_code, detail=error_data)
    
    @staticmethod
    def created(data: Any, message: Optional[str] = None) -> JSONResponse:
        """Create a 201 Created response."""
        response = {
            "success": True,
            "data": data
        }
        if message:
            response["message"] = message
        return JSONResponse(status_code=201, content=response)
    
    @staticmethod
    def accepted(data: Any, message: Optional[str] = None) -> JSONResponse:
        """Create a 202 Accepted response."""
        response = {
            "success": True,
            "data": data
        }
        if message:
            response["message"] = message
        return JSONResponse(status_code=202, content=response)


def handle_service_error(service_name: str, error: Exception) -> HTTPException:
    """Handle service-specific errors with consistent formatting."""
    error_message = f"{service_name} error: {str(error)}"
    
    # Map specific exceptions to HTTP status codes
    if isinstance(error, ValueError):
        return APIResponse.error(error_message, "VALIDATION_ERROR", status_code=400)
    elif isinstance(error, ConnectionError):
        return APIResponse.error(error_message, "SERVICE_UNAVAILABLE", status_code=503)
    elif isinstance(error, TimeoutError):
        return APIResponse.error(error_message, "TIMEOUT_ERROR", status_code=504)
    else:
        return APIResponse.error(error_message, "INTERNAL_ERROR", status_code=500)