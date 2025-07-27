Python SDK
The Pinecone Python SDK is distributed on PyPI using the package name pinecone. By default, the pinecone package has a minimal set of dependencies and interacts with Pinecone via HTTP requests. However, you can install the following extras to unlock additional functionality:

    pinecone[grpc] adds dependencies on grpcio and related libraries needed to run data operations such as upserts and queries over gRPC for a modest performance improvement.
    pinecone[asyncio] adds a dependency on aiohttp and enables usage of async methods for use with asyncio. For more details, see Asyncio support.

See the Pinecone Python SDK documentation for full installation instructions, usage examples, and reference information.To make a feature request or report an issue, please file an issue.
​
Requirements
The Pinecone Python SDK requires Python 3.9 or later. It has been tested with CPython versions from 3.9 to 3.13.
​
SDK versions
SDK versions are pinned to specific API versions. When a new API version is released, a new version of the SDK is also released. The mappings between API versions and Python SDK versions are as follows:
API version	SDK version
2025-04 (latest)	v7.x
2025-01	v6.x
2024-10	v5.3.x
2024-07	v5.0.x-v5.2.x
2024-04	v4.x
When a new stable API version is released, you should upgrade your SDK to the latest version to ensure compatibility with the latest API changes.
​
Install
To install the latest version of the Python SDK, run the following command:
Copy

# Install the latest version
pip install pinecone

# Install the latest version with gRPC extras
pip install "pinecone[grpc]"

# Install the latest version with asyncio extras
pip install "pinecone[asyncio]"

To install a specific version of the Python SDK, run the following command:
pip
Copy

# Install a specific version
pip install pinecone==<version>

# Install a specific version with gRPC extras
pip install "pinecone[grpc]"==<version>

# Install a specific version with asyncio extras
pip install "pinecone[asyncio]"==<version>

To check your SDK version, run the following command:
pip
Copy

pip show pinecone

To use the Inference API, you must be on version 5.0.0 or later.
​
Install the Pinecone Assistant Python plugin
As of Python SDK v7.0.0, the pinecone-plugin-assistant package is included by default. It is only necessary to install the package if you are using a version of the Python SDK prior to v7.0.0.
HTTP
Copy

pip install --upgrade pinecone pinecone-plugin-assistant

​
Upgrade
Before upgrading to v6.0.0, update all relevant code to account for the breaking changes explained here.Also, make sure to upgrade using the pinecone package name instead of pinecone-client; upgrading with the latter will not work as of v6.0.0.
If you already have the Python SDK, upgrade to the latest version as follows:
Copy

# Upgrade to the latest version
pip install pinecone --upgrade

# Upgrade to the latest version with gRPC extras
pip install "pinecone[grpc]" --upgrade

# Upgrade to the latest version with asyncio extras
pip install "pinecone[asyncio]" --upgrade

​
Initialize
Once installed, you can import the library and then use an API key to initialize a client instance:
Copy

from pinecone import Pinecone

pc = Pinecone(api_key="YOUR_API_KEY")

When creating an index, import the ServerlessSpec or PodSpec class as well:
Copy

from pinecone.grpc import PineconeGRPC as Pinecone
from pinecone import ServerlessSpec

pc = Pinecone(api_key="YOUR_API_KEY")

pc.create_index(
  name="docs-example",
  dimension=1536,
  metric="cosine",
  spec=ServerlessSpec(
    cloud="aws",
    region="us-east-1"
  )
)

​
Proxy configuration
If your network setup requires you to interact with Pinecone through a proxy, you will need to pass additional configuration using optional keyword parameters:

    proxy_url: The location of your proxy. This could be an HTTP or HTTPS URL depending on your proxy setup.
    proxy_headers: Accepts a python dictionary which can be used to pass any custom headers required by your proxy. If your proxy is protected by authentication, use this parameter to pass basic authentication headers with a digest of your username and password. The make_headers utility from urllib3 can be used to help construct the dictionary. Note: Not supported with Asyncio.
    ssl_ca_certs: By default, the client will perform SSL certificate verification using the CA bundle maintained by Mozilla in the certifi package. If your proxy is using self-signed certicates, use this parameter to specify the path to the certificate (PEM format).
    ssl_verify: SSL verification is enabled by default, but it is disabled when set to False. It is not recommened to go into production with SSL verification disabled.

Copy

from pinecone import Pinecone
import urllib3 import make_headers

pc = Pinecone(
    api_key="YOUR_API_KEY",
    proxy_url='https://your-proxy.com',
    proxy_headers=make_headers(proxy_basic_auth='username:password'),
    ssl_ca_certs='path/to/cert-bundle.pem'
)

​
Async requests
Pinecone Python SDK versions 6.0.0 and later provide async methods for use with asyncio. Asyncio support makes it possible to use Pinecone with modern async web frameworks such as FastAPI, Quart, and Sanic, and should significantly increase the efficiency of running requests in parallel. Use the PineconeAsyncio class to create and manage indexes and the IndexAsyncio class to read and write index data. To ensure that sessions are properly closed, use the async with syntax when creating PineconeAsyncio and IndexAsyncio objects.
Copy

# pip install "pinecone[asyncio]"
import asyncio
from pinecone import PineconeAsyncio, ServerlessSpec

async def main():
    async with PineconeAsyncio(api_key="YOUR_API_KEY") as pc:
        if not await pc.has_index(index_name):
            desc = await pc.create_index(
                name="docs-example",
                dimension=1536,
                metric="cosine",
                spec=ServerlessSpec(
                    cloud="aws",
                    region="us-east-1"
                ),
                deletion_protection="disabled",
                tags={
                    "environment": "development"
                }
            )

asyncio.run(main())

​
Query across namespaces
Each query is limited to a single namespace. However, the Pinecone Python SDK provides a query_namespaces utility method to run a query in parallel across multiple namespaces in an index and then merge the result sets into a single ranked result set with the top_k most relevant results. The query_namespaces method accepts most of the same arguments as query with the addition of a required namespaces parameter.

When using the Python SDK without gRPC extras, to get good performance, it is important to set values for the pool_threads and connection_pool_maxsize properties on the index client. The pool_threads setting is the number of threads available to execute requests, while connection_pool_maxsize is the number of cached http connections that will be held. Since these tasks are not computationally heavy and are mainly i/o bound, it should be okay to have a high ratio of threads to cpus.The combined results include the sum of all read unit usage used to perform the underlying queries for each namespace.
Python
Copy

from pinecone import Pinecone

pc = Pinecone(api_key="YOUR_API_KEY")
index = pc.Index(
    name="docs-example",
    pool_threads=50,             # <-- make sure to set these
    connection_pool_maxsize=50,  # <-- make sure to set these
)

query_vec = [ 0.1, ...] # an embedding vector with same dimension as the index
combined_results = index.query_namespaces(
    vector=query_vec,
    namespaces=['ns1', 'ns2', 'ns3', 'ns4'],
    metric="cosine",
    top_k=10,
    include_values=False,
    include_metadata=True,
    filter={"genre": { "$eq": "comedy" }},
    show_progress=False,
)

for scored_vec in combined_results.matches:
    print(scored_vec)
print(combined_results.usage)

​
Upsert from a dataframe
To quickly ingest data when using the Python SDK, use the upsert_from_dataframe method. The method includes retry logic andbatch_size, and is performant especially with Parquet file data sets. The following example upserts the uora_all-MiniLM-L6-bm25 dataset as a dataframe.
Python
Copy

from pinecone import Pinecone, ServerlessSpec
from pinecone_datasets import list_datasets, load_dataset

pc = Pinecone(api_key="API_KEY")

dataset = load_dataset("quora_all-MiniLM-L6-bm25")

pc.create_index(
  name="docs-example",
  dimension=384,
  metric="cosine",
  spec=ServerlessSpec(
    cloud="aws",
    region="us-east-1"
  )
)

# To get the unique host for an index, 
# see https://docs.pinecone.io/guides/manage-data/target-an-index
index = pc.Index(host="INDEX_HOST")

index.upsert_from_dataframe(dataset.drop(columns=["blob"]))
