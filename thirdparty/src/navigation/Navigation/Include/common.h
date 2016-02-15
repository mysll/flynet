#ifndef COMMON_H
#define COMMON_H
#include <stdint.h>
/** 安全的释放一个指针内存 */
#define SAFE_RELEASE(i)										\
	if (i)													\
		{													\
			delete i;										\
			i = NULL;										\
		}

/** 安全的释放一个指针数组内存 */
#define SAFE_RELEASE_ARRAY(i)								\
	if (i)													\
		{													\
			delete[] i;										\
			i = NULL;										\
		}


typedef int64_t													int64;
typedef int32_t													int32;
typedef int16_t													int16;
typedef int8_t													int8;
typedef uint64_t												uint64;
typedef uint32_t												uint32;
typedef uint16_t												uint16;
typedef uint8_t													uint8;
typedef uint16_t												WORD;
typedef uint32_t												DWORD;

struct Vector3{
	float x;
	float y;
	float z;
};
#endif