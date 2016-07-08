#pragma once
#include <map>
#include <string>
#include "navigation_handle.h"

class MeshNavigation {
public:
	MeshNavigation();
	~MeshNavigation();
	NavigationHandle* findNavigation(std::string resPath);
	NavigationHandle* LoadNavitagion(std::string respath, const std::map< int, std::string >& params);
	void Finalise();
private:
	std::map<std::string, NavigationHandle*> navhandles;
};
