
#ifndef NAVIGATEHANDLE_H
#define NAVIGATEHANDLE_H
#include <vector>
#include <string>
#include "common.h"

class NavigationHandle 
{
public:
	static const int NAV_ERROR = -1;

	enum NAV_TYPE
	{
		NAV_UNKNOWN = 0,
		NAV_MESH = 1,
		NAV_TILE = 2
	};

	enum NAV_OBJECT_STATE
	{
		NAV_OBJECT_STATE_MOVING = 1,	// 移动中
		NAV_OBJECT_STATE_MOVEOVER = 2,	// 移动已经结束了
	};

	NavigationHandle():
	resPath()
	{
	}

	virtual ~NavigationHandle()
	{
	}

	virtual NavigationHandle::NAV_TYPE type() const{ return NAV_UNKNOWN; }

	virtual int findStraightPath(int layer, const Vector3 & start, const Vector3 & end, std::vector<Vector3>& paths) = 0;
	virtual int raycast(int layer, const Vector3 & start, const Vector3 & end, Vector3 & hitPointVec) = 0;

	std::string resPath;
};

#endif

