// Weak fallbacks for the --wrap=malloc family in the esp32 and esp32s3 targets.
// espradio provides strong wrappers when it is linked.

typedef __SIZE_TYPE__ size_t;

void *__real_malloc(size_t size);
void *__real_calloc(size_t n, size_t size);
void *__real_realloc(void *ptr, size_t size);
void __real_free(void *ptr);

__attribute__((weak)) void *__wrap_malloc(size_t size) {
	return __real_malloc(size);
}

__attribute__((weak)) void *__wrap_calloc(size_t n, size_t size) {
	return __real_calloc(n, size);
}

__attribute__((weak)) void *__wrap_realloc(void *ptr, size_t size) {
	return __real_realloc(ptr, size);
}

__attribute__((weak)) void __wrap_free(void *ptr) {
	__real_free(ptr);
}
