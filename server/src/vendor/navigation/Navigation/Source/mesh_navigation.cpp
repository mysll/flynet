#include <stdio.h>
#include <common.h>
#include "mesh_navigation.h"
#include "navigation_mesh_handle.h"
#include "navigation_tile_handle.h"

MeshNavigation::MeshNavigation()
{

}

MeshNavigation::~MeshNavigation()
{

}

void MeshNavigation::Finalise()
{
	std::map<std::string, NavigationHandle*>::iterator iter = navhandles.begin();
	for(; iter != navhandles.end(); ++iter) {
		delete iter->second;
	}

	navhandles.clear();
}

NavigationHandle* MeshNavigation::findNavigation(std::string resPath)
{
	std::map<std::string, NavigationHandle*>::iterator iter = navhandles.find(resPath);
	if(iter != navhandles.end())
	{
		if(iter->second == NULL)
			return NULL;

		if(iter->second->type() == NavigationHandle::NAV_MESH)
		{
			return iter->second;
		}

	}

	return NULL;
}

NavigationHandle* MeshNavigation::LoadNavitagion(std::string resPath, const std::map< int, std::string >& params, int type)
{
	if(resPath == "")
		return NULL;
	
	std::map<std::string, NavigationHandle*>::iterator iter = navhandles.find(resPath);
	if(iter != navhandles.end())
	{
		return iter->second;
	}

	NavigationHandle* pNavigationHandle_ = NULL;
	if(type == 1)
		pNavigationHandle_ = NavTileHandle::create(resPath, params);
	else if(type == 2)
		pNavigationHandle_ = NavMeshHandle::create(resPath, params);
	if(pNavigationHandle_ != NULL)
		navhandles[resPath] = pNavigationHandle_;
	return pNavigationHandle_;
}

