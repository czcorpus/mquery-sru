// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//   This file is part of MQUERY.
//
//  MQUERY is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  MQUERY is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with MQUERY.  If not, see <https://www.gnu.org/licenses/>.

#ifdef __cplusplus
extern "C" {
#endif

typedef void* ConcV;

typedef void* KWICRowsV;

typedef long long int PosInt;

typedef struct ConcRetval {
    ConcV value;
    const char * err;
} ConcRetval;

typedef struct KWICRowsRetval {
    KWICRowsV value;
    PosInt size;
    PosInt concSize;
    const char * err;
    int errorCode;
} KWICRowsRetval;


/**
 * @brief Based on provided query, return at most `limit` sentences matching the query.
 * Please note that when called from Go via function `GetConcExamples`, the Go function
 * checks the `limit` argument against `mango.MaxRecordsInternalLimit` and will not allow
 * larger value.
 *
 * @param corpusPath
 * @param query
 * @param attrs Positional attributes (comma-separated) to be attached to returned tokens
 * @param limit
 * @return KWICRowsRetval
 */
KWICRowsRetval conc_examples(
    const char* corpusPath, const char*query, const char* attrs, PosInt fromLine, PosInt limit);


/**
 * @brief This function frees all the allocated memory
 * for a concordance example. It is intended to be called
 * from Go.
 *
 * @param value
 * @param numItems
 */
void conc_examples_free(KWICRowsV value, int numItems);


#ifdef __cplusplus
}
#endif