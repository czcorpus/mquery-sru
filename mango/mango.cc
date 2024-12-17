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


#include "corp/corpus.hh"
#include "concord/concord.hh"
#include "concord/concget.hh"
#include "query/cqpeval.hh"
#include "mango.h"
#include <cmath>

using namespace std;


/**
 * @brief Based on provided query, return at most `limit` sentences matching the query.
 *
 * @param corpusPath
 * @param query
 * @param attrs Positional attributes (comma-separated) to be attached to returned tokens
 * @param limit
 * @return KWICRowsRetval
 */
KWICRowsRetval conc_examples(
    const char* corpusPath,
    const char* query,
    const char* attrs,
    const char* structs,
    const char* refs,
    const char* refsSplitter,
    PosInt fromLine,
    PosInt limit,
    PosInt maxContext,
    const char* viewContextStruct) {

    string cPath(corpusPath);
    try {
        Corpus* corp = new Corpus(cPath);
        Concordance* conc = new Concordance(
            corp, corp->filter_query(eval_cqpquery(query, corp)));
        conc->sync();
        if (conc->size() == 0 && fromLine == 0) {
            KWICRowsRetval ans {
                nullptr,
                0,
                0,
                nullptr
            };
            return ans;
        }
        if (conc->size() < fromLine) {
            const char* msg = "line range out of result size";
            char* dynamicStr = static_cast<char*>(malloc(strlen(msg) + 1));
            strcpy(dynamicStr, msg);
            KWICRowsRetval ans {
                nullptr,
                0,
                0,
                dynamicStr,
                1
            };
            return ans;
        }
        conc->shuffle();
        PosInt concSize = conc->size();
        std::string cppContextStruct(viewContextStruct);
        std::string halfLeft = "-" + std::to_string(int(std::floor(maxContext / 2.0)));
        std::string halfRight = std::to_string(int(std::ceil(maxContext / 2.0)));
        KWICLines* kl = new KWICLines(
            corp,
            conc->RS(true, fromLine, fromLine+limit),
            cppContextStruct.empty() ? halfLeft.c_str() : ("-1:"+cppContextStruct).c_str(),
            cppContextStruct.empty() ? halfRight.c_str() : ("1:"+cppContextStruct).c_str(),
            attrs,
            attrs,
            structs,
            refs,
            maxContext,
            false
        );
        if (conc->size() < limit) {
            limit = conc->size();
        }
        char** lines = (char**)malloc(limit * sizeof(char*));
        int i = 0;
        while (kl->nextline()) {
            auto lft = kl->get_left();
            auto kwc = kl->get_kwic();
            auto rgt = kl->get_right();
            std::ostringstream buffer;

            buffer << kl->get_refs() << refsSplitter;

            for (size_t i = 0; i < lft.size(); ++i) {
                if (i > 0) {
                    buffer << " ";
                }
                buffer << lft.at(i);
            }
            for (size_t i = 0; i < kwc.size(); ++i) {
                if (i > 0) {
                    buffer << " ";
                }
                buffer << kwc.at(i);
            }
            for (size_t i = 0; i < rgt.size(); ++i) {
                if (i > 0) {
                    buffer << " ";
                }
                buffer << rgt.at(i);
            }
            lines[i] = strdup(buffer.str().c_str());
            i++;
            if (i == limit) {
                break;
            }
        }
        // We've allocated memory for `limit` rows,
        // but it's possible that there is less rows
        // available so here we fill the remaining items
        // with empty strings.
        for (int i2 = i; i2 < limit; i2++) {
            lines[i2] = strdup("");
        }
        delete conc;
        delete corp;
        KWICRowsRetval ans {
            lines,
            limit,
            concSize,
            nullptr,
            0
        };
        return ans;

    } catch (std::exception &e) {
        KWICRowsRetval ans {
            nullptr,
            0,
            0,
            strdup(e.what()),
            0
        };
        return ans;
    }
}

void conc_examples_free(KWICRowsV value, int numItems) {
    char** tValue = (char**)value;
    for (int i = 0; i < numItems; i++) {
        free(tValue[i]);
    }
    free(tValue);
}
