#pragma once
#ifdef __cplusplus
extern "C" {
#endif
int InitNavigation();
int DestroyNavigation();
int GetPathArrSize(const float * paths);
int CreateNavigation(int mapid, const char * path, const char * file);
float* FindStraightPath(int mapid, int layer, float startx, float starty, float startz, float endx, float endy, float endz);
float*  Raycast(int mapid, int layer, float startx, float starty, float startz, float endx, float endy, float endz);
float GetHeight(int mapid, int layer, float x, float y, float z);
void FreePaths(float * paths);
void Free(void * ptr);
#ifdef __cplusplus    
}
#endif