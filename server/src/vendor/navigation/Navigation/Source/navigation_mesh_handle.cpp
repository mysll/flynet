#include <stdio.h>
#include <assert.h>
#include <memory.h>
#include "common.h"
#include "navigation_mesh_handle.h"

//-------------------------------------------------------------------------------------
NavMeshHandle::NavMeshHandle():
NavigationHandle(),
navmeshLayer()
{
}

//-------------------------------------------------------------------------------------
NavMeshHandle::~NavMeshHandle()
{
	std::map<int, NavmeshLayer>::iterator iter = navmeshLayer.begin();
	for(; iter != navmeshLayer.end(); ++iter)
	{
		dtFreeNavMesh(iter->second.pNavmesh);
		dtFreeNavMeshQuery(iter->second.pNavmeshQuery);
	}
	
	printf("NavMeshHandle::~NavMeshHandle(): (%s) is destroyed!\n", resPath.c_str());
}

//-------------------------------------------------------------------------------------
int NavMeshHandle::findStraightPath(int layer, const Vector3 & start, const Vector3 & end, std::vector<Vector3>& paths)
{
	std::map<int, NavmeshLayer>::iterator iter = navmeshLayer.find(layer);
	if(iter == navmeshLayer.end())
	{
		printf("NavMeshHandle::findStraightPath: not found layer(%d)\n",  layer);
		return NAV_ERROR;
	}

	dtNavMeshQuery* navmeshQuery = iter->second.pNavmeshQuery;
	// dtNavMesh* 

	float spos[3];
	spos[0] = start.x;
	spos[1] = start.y;
	spos[2] = start.z;

	float epos[3];
	epos[0] = end.x;
	epos[1] = end.y;
	epos[2] = end.z;

	dtQueryFilter filter;
	filter.setIncludeFlags(0xffff);
	filter.setExcludeFlags(0);

	const float extents[3] = {2.f, 4.f, 2.f};

	dtPolyRef startRef = INVALID_NAVMESH_POLYREF;
	dtPolyRef endRef = INVALID_NAVMESH_POLYREF;

	float startNearestPt[3];
	float endNearestPt[3];
	navmeshQuery->findNearestPoly(spos, extents, &filter, &startRef, startNearestPt);
	navmeshQuery->findNearestPoly(epos, extents, &filter, &endRef, endNearestPt);

	if (!startRef || !endRef)
	{
		printf("NavMeshHandle::findStraightPath(%s): Could not find any nearby poly's (%d, %d)\n",resPath.c_str(), startRef, endRef );
		return NAV_ERROR_NEARESTPOLY;
	}

	dtPolyRef polys[MAX_POLYS];
	int npolys;
	float straightPath[MAX_POLYS * 3];
	unsigned char straightPathFlags[MAX_POLYS];
	dtPolyRef straightPathPolys[MAX_POLYS];
	int nstraightPath;
	int pos = 0;

	navmeshQuery->findPath(startRef, endRef, startNearestPt, endNearestPt, &filter, polys, &npolys, MAX_POLYS);
	nstraightPath = 0;

	if (npolys)
	{
		float epos1[3];
		bool posOverPoly;
		dtVcopy(epos1, endNearestPt);
				
		if (polys[npolys-1] != endRef)
			navmeshQuery->closestPointOnPoly(polys[npolys-1], endNearestPt, epos1, &posOverPoly);
				
		navmeshQuery->findStraightPath(startNearestPt, endNearestPt, polys, npolys, straightPath, straightPathFlags, straightPathPolys, &nstraightPath, MAX_POLYS);

		Vector3 currpos;
		for(int i = 0; i < nstraightPath * 3; )
		{
			currpos.x = straightPath[i++];
			currpos.y = straightPath[i++];
			currpos.z = straightPath[i++];
			paths.push_back(currpos);
			pos++; 
			
			//DEBUG_MSG(fmt::format("NavMeshHandle::findStraightPath: {}->{}, {}, {}\n", pos, currpos.x, currpos.y, currpos.z));
		}
	}

	return pos;
}

//-------------------------------------------------------------------------------------
int NavMeshHandle::raycast(int layer, const Vector3 & start, const Vector3 & end, Vector3 & hitPointVec)
{
	std::map<int, NavmeshLayer>::iterator iter = navmeshLayer.find(layer);
	if(iter == navmeshLayer.end())
	{
		printf("NavMeshHandle::raycast: not found layer(%d)\n",  layer);
		return NAV_ERROR;
	}

	dtNavMeshQuery* navmeshQuery = iter->second.pNavmeshQuery;

	float hitPoint[3];

	float spos[3];
	spos[0] = start.x;
	spos[1] = start.y;
	spos[2] = start.z;

	float epos[3];
	epos[0] = end.x;
	epos[1] = end.y;
	epos[2] = end.z;

	dtQueryFilter filter;
	filter.setIncludeFlags(0xffff);
	filter.setExcludeFlags(0);

	const float extents[3] = {2.f, 4.f, 2.f};

	dtPolyRef startRef = INVALID_NAVMESH_POLYREF;

	float nearestPt[3];
	navmeshQuery->findNearestPoly(spos, extents, &filter, &startRef, nearestPt);

	if (!startRef)
	{
		return NAV_ERROR_NEARESTPOLY;
	}

	float t = 0;
	float hitNormal[3];
	memset(hitNormal, 0, sizeof(hitNormal));

	dtPolyRef polys[MAX_POLYS];
	int npolys;

	navmeshQuery->raycast(startRef, spos, epos, &filter, &t, hitNormal, polys, &npolys, MAX_POLYS);

	if (t > 1)
	{
		// no hit
		return NAV_ERROR;
	}
	else
	{
		// Hit
		hitPoint[0] = spos[0] + (epos[0] - spos[0]) * t;
		hitPoint[1] = spos[1] + (epos[1] - spos[1]) * t;
		hitPoint[2] = spos[2] + (epos[2] - spos[2]) * t;
		if (npolys)
		{
			float h = 0;
			navmeshQuery->getPolyHeight(polys[npolys-1], hitPoint, &h);
			hitPoint[1] = h;
		}
	}
	hitPointVec.x = hitPoint[0];
	hitPointVec.y = hitPoint[1];
	hitPointVec.z = hitPoint[2];
	return 0;
}

//-------------------------------------------------------------------------------------
NavigationHandle* NavMeshHandle::create(std::string resPath, const std::map< int, std::string >& params)
{
	if(resPath == "")
		return NULL;
	
	NavMeshHandle* pNavMeshHandle = NULL;

	std::string path = resPath;

	if(params.size() == 0)
	{	
		printf("NavMeshHandle::create: not found navmesh.!\n");
		return NULL;
	}
	else
	{
		pNavMeshHandle = new NavMeshHandle();
		std::map< int, std::string >::const_iterator iter = params.begin();

		for(; iter != params.end(); ++iter)
		{
			_create(iter->first, resPath, path + "/" + iter->second, pNavMeshHandle);
		}		
	}
	
	return pNavMeshHandle;
}

//-------------------------------------------------------------------------------------
bool NavMeshHandle::_create(int layer, const std::string& resPath, const std::string& res, NavMeshHandle* pNavMeshHandle)
{
	assert(pNavMeshHandle != NULL);
	FILE* fp = fopen(res.c_str(), "rb");
	if (!fp)
	{
		printf("NavMeshHandle::create: open(%s) is error!\n", 
			res.c_str());

		return false;
	}
	
	printf("NavMeshHandle::create: (%s), layer=%d\n", 
		res.c_str(), layer);

	bool safeStorage = true;
	int pos = 0;
	int size = sizeof(NavMeshSetHeader);
	
	fseek(fp, 0, SEEK_END); 
	size_t flen = ftell(fp); 
	fseek(fp, 0, SEEK_SET); 

	uint8* data = new uint8[flen];
	if(data == NULL)
	{
		printf("NavMeshHandle::create: open(%s), memory(size=%d) error!\n", 
			res.c_str(), flen);

		fclose(fp);
		SAFE_RELEASE_ARRAY(data);
		return false;
	}

	size_t readsize = fread(data, 1, flen, fp);
	if(readsize != flen)
	{
		printf("NavMeshHandle::create: open(%s), read(size=%d != %d) error!\n", 
			res.c_str(), readsize, flen);

		fclose(fp);
		SAFE_RELEASE_ARRAY(data);
		return false;
	}

	if (readsize < sizeof(NavMeshSetHeader))
	{
		printf("NavMeshHandle::create: open(%s), NavMeshSetHeader is error!\n", 
			res.c_str());

		fclose(fp);
		SAFE_RELEASE_ARRAY(data);
		return false;
	}

	NavMeshSetHeader header;
	memcpy(&header, data, size);

	pos += size;

	if (header.version != NavMeshHandle::RCN_NAVMESH_VERSION)
	{
		printf("NavMeshHandle::create: navmesh version(%d) is not match(%d)!\n", 
			header.version, ((int)NavMeshHandle::RCN_NAVMESH_VERSION));

		fclose(fp);
		SAFE_RELEASE_ARRAY(data);
		return false;
	}

	dtNavMesh* mesh = dtAllocNavMesh();
	if (!mesh)
	{
		printf("NavMeshHandle::create: dtAllocNavMesh is failed!\n");
		fclose(fp);
		SAFE_RELEASE_ARRAY(data);
		return false;
	}

	dtStatus status = mesh->init(&header.params);
	if (dtStatusFailed(status))
	{
		printf("NavMeshHandle::create: mesh init is error(%d)!\n", status);
		fclose(fp);
		SAFE_RELEASE_ARRAY(data);
		return false;
	}

	// Read tiles.
	bool success = true;
	for (int i = 0; i < header.tileCount; ++i)
	{
		NavMeshTileHeader tileHeader;
		size = sizeof(NavMeshTileHeader);
		memcpy(&tileHeader, &data[pos], size);
		pos += size;

		size = tileHeader.dataSize;
		if (!tileHeader.tileRef || !tileHeader.dataSize)
		{
			success = false;
			status = DT_FAILURE + DT_INVALID_PARAM;
			break;
		}
		
		unsigned char* tileData = 
			(unsigned char*)dtAlloc(size, DT_ALLOC_PERM);
		if (!tileData)
		{
			success = false;
			status = DT_FAILURE + DT_OUT_OF_MEMORY;
			break;
		}
		memcpy(tileData, &data[pos], size);
		pos += size;

		status = mesh->addTile(tileData
			, size
			, (safeStorage ? DT_TILE_FREE_DATA : 0)
			, tileHeader.tileRef
			, 0);

		if (dtStatusFailed(status))
		{
			success = false;
			break;
		}
	}

	fclose(fp);
	SAFE_RELEASE_ARRAY(data);

	if (!success)
	{
		printf("NavMeshHandle::create:  error(%d)!\n", status);
		dtFreeNavMesh(mesh);
		return false;
	}

	dtNavMeshQuery* pMavmeshQuery = new dtNavMeshQuery();

	pMavmeshQuery->init(mesh, 1024);
	pNavMeshHandle->resPath = resPath;
	pNavMeshHandle->navmeshLayer[layer].pNavmeshQuery = pMavmeshQuery;
	pNavMeshHandle->navmeshLayer[layer].pNavmesh = mesh;
	
	uint32 tileCount = 0;
	uint32 nodeCount = 0;
	uint32 polyCount = 0;
	uint32 vertCount = 0;
	uint32 triCount = 0;
	uint32 triVertCount = 0;
	uint32 dataSize = 0;

	const dtNavMesh* navmesh = mesh;
	for (int32 i = 0; i < navmesh->getMaxTiles(); ++i)
	{
		const dtMeshTile* tile = navmesh->getTile(i);
		if (!tile || !tile->header)
			continue;

		tileCount ++;
		nodeCount += tile->header->bvNodeCount;
		polyCount += tile->header->polyCount;
		vertCount += tile->header->vertCount;
		triCount += tile->header->detailTriCount;
		triVertCount += tile->header->detailVertCount;
		dataSize += tile->dataSize;

		// DEBUG_MSG(fmt::format("NavMeshHandle::create: verts({}, {}, {})\n", tile->verts[0], tile->verts[1], tile->verts[2]));
	}

	printf("\t==> tiles loaded: %d\n", tileCount);
	printf("\t==> BVTree nodes: %d\n", nodeCount);
	printf("\t==> %d polygons (%d vertices)\n", polyCount, vertCount);
	printf("\t==> %d triangles (%d vertices)\n", triCount, triVertCount);
	printf("\t==> %.2f MB of data (not including pointers)\n", (((float)dataSize / sizeof(unsigned char)) / 1048576));
	
	return true;
}
