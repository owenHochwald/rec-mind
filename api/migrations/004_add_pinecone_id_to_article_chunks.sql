-- Add pinecone_id field to article_chunks table
-- This field will store the Pinecone vector ID for each chunk

ALTER TABLE article_chunks 
ADD COLUMN pinecone_id VARCHAR(255) NULL;

-- Add index on pinecone_id for faster lookups
CREATE INDEX idx_article_chunks_pinecone_id ON article_chunks(pinecone_id);

-- Add comment for clarity
COMMENT ON COLUMN article_chunks.pinecone_id IS 'Reference to the Pinecone vector ID for this chunk';

-- Verify the table structure after migration
-- Expected columns: id, article_id, chunk_index, content, token_count, character_count, created_at, pinecone_id