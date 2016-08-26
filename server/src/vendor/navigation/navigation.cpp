#include <stdio.h>
#include <string>
#include <map>
#include <vector>
#include <stdlib.h>
#include "navigation.h"
#include "mesh_navigation.h"
#include "common.h"

MeshNavigation* g_Navigation;
std::map<int, NavigationHandle*> g_navHandles;
int InitNavigation() {

	g_Navigation = new MeshNavigation();
	return 0;
}

int CreateNavigation(int mapid, const char * path, const char * file, int type) {
	if(g_Navigation == NULL)
		return -1;
	if(g_navHandles.find(mapid) != g_navHandles.end()) {
		return -1;
	}

	std::string respath = std::string(path);
	NavigationHandle* ptr = g_Navigation->findNavigation(respath);
	if(ptr != NULL) {
		g_navHandles[mapid] = ptr;
		return 0;
	}

	std::map<int, std::string> params;
	params[0] = std::string(file);
	ptr = g_Navigation->LoadNavitagion(path, params, type);
	if(ptr == NULL)
		return -1;

	g_navHandles[mapid] = ptr;

	return 0;
}

void Free(void * ptr) {
	if(ptr) {
		free(ptr);
		ptr = NULL;
	}
}

int DestroyNavigation() {
	g_Navigation->Finalise();
	g_navHandles.clear();
	delete g_Navigation;
	return 0;
}

int GetPathArrSize(const float* paths) {
	if(paths == NULL)
		return 0;
	return (int)paths[0];
}

void FreePaths(float * paths) {
	SAFE_RELEASE_ARRAY(paths);
}

float * FindStraightPath(int mapid, int layer, float startx, float starty, float startz, float endx, float endy, float endz)
{
	Vector3 start(startx, starty, startz);
	Vector3 stop(endx, endy, endz);
	std::vector<Vector3> paths;
	std::map<int, NavigationHandle*>::iterator iter1 = g_navHandles.find(mapid);
	if(iter1 == g_navHandles.end() ) {
		return NULL;
	}

	int ret = iter1->second->findStraightPath(0, start, stop, paths);
	if(ret <= 0)
		return NULL;

	float *patharr = (float*)malloc(sizeof(float) * ret * 3 + 1);
	if(patharr == NULL) {
		printf("out of memory");
		return NULL;
	}
	patharr[0] = float(ret*3); //第一位保存长度
	int index = 1;
	std::vector<Vector3>::iterator iter = paths.begin();
	for(; iter != paths.end(); ++iter)
	{
		patharr[index++] = iter->x;
		patharr[index++] = iter->y;
		patharr[index++] = iter->z;
	}

	return patharr;
}

float* Raycast(int mapid, int layer, float startx, float starty, float startz, float endx, float endy, float endz)
{
	Vector3 start(startx, starty, startz);
	Vector3 stop(endx, endy, endz);
	std::map<int, NavigationHandle*>::iterator iter1 = g_navHandles.find(mapid);
	if(iter1 == g_navHandles.end() ) {
		return NULL;
	}
	std::vector<Vector3>  hitPointVec;
	int ret = iter1->second->raycast(layer, start, stop, hitPointVec);
	if(ret <= 0)
		return NULL;

	float *retpos = (float*)malloc(sizeof(float) * ret * 3 + 1);
	if(retpos == NULL) {
		printf("out of memory");
		return NULL;
	}

	retpos[0] = float(ret*3); //第一位保存长度
	int index = 1;
	std::vector<Vector3>::iterator iter = hitPointVec.begin();
	for(; iter != hitPointVec.end(); ++iter)
	{
		retpos[index++] = iter->x;
		retpos[index++] = iter->y;
		retpos[index++] = iter->z;
	}

	return retpos;
}

float GetHeight(int mapid, int layer, float x, float y, float z)
{
	return 0;
}