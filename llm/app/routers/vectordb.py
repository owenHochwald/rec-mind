"""Vector database endpoints."""

from fastapi import APIRouter, Depends

from ..vectordb import get_vectordb_service, VectorDBService
from ..utils import endpoint_error_handler

router = APIRouter()


@router.get("/index/stats")
@endpoint_error_handler("index statistics retrieval")
async def get_index_stats(
    vectordb_svc: VectorDBService = Depends(get_vectordb_service)
):
    """Get Pinecone index statistics."""
    return await vectordb_svc.get_index_stats()