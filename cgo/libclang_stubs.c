// This file implements small trampoline functions around libclang. Structs
// larger than 16 bytes are passed and returned through pointers, because TinyGo
// does not lower them to the C ABI yet. See:
// https://github.com/tinygo-org/tinygo/issues/4489

#include <clang-c/Index.h> // If this fails, libclang headers aren't available. Please take a look here: https://tinygo.org/docs/guides/build/

typedef int (*tinygo_clang_visitor)(CXCursor *c, CXCursor *parent, CXClientData client_data);

struct tinygo_clang_visitor_data {
	tinygo_clang_visitor visitor;
	CXClientData client_data;
};

static enum CXChildVisitResult tinygo_clang_visitor_trampoline(CXCursor c, CXCursor parent, CXClientData client_data) {
	struct tinygo_clang_visitor_data *data = client_data;
	return data->visitor(&c, &parent, data->client_data);
}

void tinygo_clang_getTranslationUnitCursor(CXTranslationUnit tu, CXCursor *result) {
	*result = clang_getTranslationUnitCursor(tu);
}

unsigned tinygo_clang_visitChildren(const CXCursor *parent, tinygo_clang_visitor visitor, CXClientData client_data) {
	struct tinygo_clang_visitor_data data = {visitor, client_data};
	return clang_visitChildren(*parent, tinygo_clang_visitor_trampoline, &data);
}

CXString tinygo_clang_getCursorSpelling(const CXCursor *c) {
	return clang_getCursorSpelling(*c);
}

CXString tinygo_clang_getCursorPrettyPrinted(const CXCursor *c, CXPrintingPolicy policy) {
	return clang_getCursorPrettyPrinted(*c, policy);
}

CXPrintingPolicy tinygo_clang_getCursorPrintingPolicy(const CXCursor *c) {
	return clang_getCursorPrintingPolicy(*c);
}

enum CXCursorKind tinygo_clang_getCursorKind(const CXCursor *c) {
	return clang_getCursorKind(*c);
}

void tinygo_clang_getCursorType(const CXCursor *c, CXType *result) {
	*result = clang_getCursorType(*c);
}

void tinygo_clang_getTypeDeclaration(const CXType *t, CXCursor *result) {
	*result = clang_getTypeDeclaration(*t);
}

void tinygo_clang_getTypedefDeclUnderlyingType(const CXCursor *c, CXType *result) {
	*result = clang_getTypedefDeclUnderlyingType(*c);
}

void tinygo_clang_getCursorResultType(const CXCursor *c, CXType *result) {
	*result = clang_getCursorResultType(*c);
}

int tinygo_clang_Cursor_getNumArguments(const CXCursor *c) {
	return clang_Cursor_getNumArguments(*c);
}

void tinygo_clang_Cursor_getArgument(const CXCursor *c, unsigned i, CXCursor *result) {
	*result = clang_Cursor_getArgument(*c, i);
}

enum CX_StorageClass tinygo_clang_Cursor_getStorageClass(const CXCursor *c) {
	return clang_Cursor_getStorageClass(*c);
}

void tinygo_clang_getCursorLocation(const CXCursor *c, CXSourceLocation *result) {
	*result = clang_getCursorLocation(*c);
}

void tinygo_clang_getCursorExtent(const CXCursor *c, CXSourceRange *result) {
	*result = clang_getCursorExtent(*c);
}

CXTranslationUnit tinygo_clang_Cursor_getTranslationUnit(const CXCursor *c) {
	return clang_Cursor_getTranslationUnit(*c);
}

long long tinygo_clang_getEnumConstantDeclValue(const CXCursor *c) {
	return clang_getEnumConstantDeclValue(*c);
}

void tinygo_clang_getEnumDeclIntegerType(const CXCursor *c, CXType *result) {
	*result = clang_getEnumDeclIntegerType(*c);
}

unsigned tinygo_clang_Cursor_isAnonymous(const CXCursor *c) {
	return clang_Cursor_isAnonymous(*c);
}

unsigned tinygo_clang_Cursor_isBitField(const CXCursor *c) {
	return clang_Cursor_isBitField(*c);
}

unsigned tinygo_clang_Cursor_isMacroFunctionLike(const CXCursor *c) {
	return clang_Cursor_isMacroFunctionLike(*c);
}

void tinygo_clang_getDiagnosticLocation(CXDiagnostic diagnostic, CXSourceLocation *result) {
	*result = clang_getDiagnosticLocation(diagnostic);
}

void tinygo_clang_tokenize(CXTranslationUnit tu, const CXSourceRange *range, CXToken **tokens, unsigned *numTokens) {
	clang_tokenize(tu, *range, tokens, numTokens);
}

CXString tinygo_clang_getTokenSpelling(CXTranslationUnit tu, const CXToken *token) {
	return clang_getTokenSpelling(tu, *token);
}

void tinygo_clang_getTokenLocation(CXTranslationUnit tu, const CXToken *token, CXSourceLocation *result) {
	*result = clang_getTokenLocation(tu, *token);
}

void tinygo_clang_getExpansionLocation(const CXSourceLocation *location, CXFile *file, unsigned *line, unsigned *column, unsigned *offset) {
	clang_getExpansionLocation(*location, file, line, column, offset);
}

void tinygo_clang_getFileLocation(const CXSourceLocation *location, CXFile *file, unsigned *line, unsigned *column, unsigned *offset) {
	clang_getFileLocation(*location, file, line, column, offset);
}

void tinygo_clang_getPresumedLocation(const CXSourceLocation *location, CXString *filename, unsigned *line, unsigned *column) {
	clang_getPresumedLocation(*location, filename, line, column);
}

unsigned tinygo_clang_isFunctionTypeVariadic(const CXType *t) {
	return clang_isFunctionTypeVariadic(*t);
}

void tinygo_clang_getArgType(const CXType *t, unsigned i, CXType *result) {
	*result = clang_getArgType(*t, i);
}

CXString tinygo_clang_getTypedefName(const CXType *t) {
	return clang_getTypedefName(*t);
}

CXString tinygo_clang_getTypeSpelling(const CXType *t) {
	return clang_getTypeSpelling(*t);
}

void tinygo_clang_Type_getNamedType(const CXType *t, CXType *result) {
	*result = clang_Type_getNamedType(*t);
}

void tinygo_clang_getCanonicalType(const CXType *t, CXType *result) {
	*result = clang_getCanonicalType(*t);
}

void tinygo_clang_getPointeeType(const CXType *t, CXType *result) {
	*result = clang_getPointeeType(*t);
}

void tinygo_clang_getElementType(const CXType *t, CXType *result) {
	*result = clang_getElementType(*t);
}

long long tinygo_clang_getArraySize(const CXType *t) {
	return clang_getArraySize(*t);
}

long long tinygo_clang_Type_getSizeOf(const CXType *t) {
	return clang_Type_getSizeOf(*t);
}

long long tinygo_clang_Type_getAlignOf(const CXType *t) {
	return clang_Type_getAlignOf(*t);
}

long long tinygo_clang_Type_getOffsetOf(const CXType *t, const char *s) {
	return clang_Type_getOffsetOf(*t, s);
}
