//implementation of tf-idf (term frequency–inverse document frequency) ranking
//
// <wikipedia> term frequency–inverse document frequency, is a numerical statistic that is intended to reflect how important a word is to a document 
// in a collection or corpus.[1]:8 It is often used as a weighting factor in information retrieval and text mining. The tf-idf value 
// increases proportionally to the number of times a word appears in the document, but is offset by the frequency of the word in the 
// corpus, which helps to adjust for the fact that some words appear more frequently in general.
// Variations of the tf–idf weighting scheme are often used by search engines as a central tool in scoring and ranking a document's 
// relevance given a user query. tf–idf can be successfully used for stop-words filtering in various subject fields including text 
// summarization and classification.

package main
import "fmt"
import "strings"
import "os"
import "log"
import "bufio"
import "path"
import "regexp"
import . "math"

// Our Term Map, term (string) to map of docid (int) to term raw  count (int)
var TermMap map[string] map[int] *int

// TermInDocs term (string) to number of docs
var TermInDocsMap map[string] *int

// DocWordCount docid (int) to words in doc 
var DocWordCount map[int] int

// DocIDs - id for each doc in our set, convieniently can be used to create input and output files
var docids [5]int

// clean_term()
// clean up each term by removing punctuation, rendering lower case, demoting plural to singular
func clean_term(term string) string{

    // clean it all up (borrowed from github.com/kennygrant/sanitize)
    name := strings.ToLower(term)
    name = path.Clean(path.Base(name))
    name = strings.Trim(name, " ")

    // remove the plural s (oversimplification)
    term = strings.TrimSuffix(term, "'s")

    // TODO: add with and without trailing s 
    term = strings.TrimSuffix(term, "s")

    // Replace certain joining characters with a dash
    seps, err := regexp.Compile(`[ &_=+:]`)
    if err == nil {
        // TODO: for hyphenation string1-string2, add both strings or "string1 string2"
        name = seps.ReplaceAllString(name, "-")
    }

    // Remove all other unrecognised characters - NB we do allow any printable characters
    legal, err := regexp.Compile(`[^[:alnum:]-]`)
    if err == nil {
        name = legal.ReplaceAllString(name, "")
    }

    // Remove any double dashes caused by existing - in name
    name = strings.Replace(name, "--", "-", -1)
    term = strings.TrimSuffix(term, "-")

    // Remove trailing comma
    name2 := name
    name2 = strings.TrimRight(name2, ",")
    return name2
}

func increment(xPtr *int) {
    if (xPtr != nil) {
        *xPtr += 1
    }
}

func add_term(term string, docid int) int {
   retval := 0

   // if new term
   if TermMap[term] == nil {
       tcount := new(int)
       increment(tcount)
       mm := make(map[int]*int)
       TermMap[term] = mm
       TermMap[term][docid] = tcount
       retval = 1
   } 

   // term is present for another docid
   if TermMap[term][docid] == nil {
       acount := new(int)
       increment(acount)
       TermMap[term][docid] = acount
   } else { // term is present for this docid, just increment
       gm := TermMap[term]
       tcount_ptr := gm[docid]
       increment(tcount_ptr)
       if (tcount_ptr != nil) {
       }
   }
   return retval
}

func add_term_doc_count(term string) int {
   retval := 0

   // if new term
   if TermInDocsMap[term] == nil {
       tcount := new(int)
       increment(tcount)
       TermInDocsMap[term] = tcount
       retval = 1
   } else {
       newcount := TermInDocsMap[term]
       increment(newcount)
   }
   return retval
}

// parse the file into terms, clean up the terms and add to our dictionary
func add_terms_to_dictionary(docid int) {
    term_count := 0

    // Open the file
    filename := fmt.Sprintf("doc%d.txt",docid)
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    } 

    // Scan the file for words
    scanner := bufio.NewScanner(file)

    // Set the split function for the scanning operation.
    scanner.Split(bufio.ScanWords)

    for scanner.Scan() {
        // make lower case
        term := strings.ToLower(scanner.Text())
       cterm := clean_term(term);
        // add words to dictionary
        term_count += add_term(cterm, docid)
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "reading input:", err)
    }
    // save the term_count: number of unique terms in the doc
    DocWordCount[docid] = term_count
}

// sum how many docs each term appears in
func sum_of_docs_per_term () {
    for term, docmap := range TermMap {
        for _, tcount := range docmap {
            add_term_doc_count(term) // how many docs have this term
            tcount = tcount // complains about not using anything
        }
    }
}

// fill in the scores and write to .csv
func write_tfidf_values(docid int) {
    var N int
    var df_t int
    var tf_td int

    // Open the file
    filename := fmt.Sprintf("doc%d.csv",docid)
    file, err := os.Create(filename)
    if err != nil {
        log.Fatal(err)
    }
    // Assemble the scores
    N = len(docids)
    
    for term, docmap := range TermMap {
        df_t = 0
        tf_td = 0
        var score float64

        df_t = *(TermInDocsMap[term])

        for doc, tcount := range docmap {
            if doc == docid {
                tf_td = *tcount
                if  df_t > 0  {
                    score = 1+ Log10( float64(tf_td)) * Log10(float64(N)/float64(df_t))
                }
            }
        }

        if (tf_td > 0) {
            term_line := fmt.Sprintf("%s,%f\n",term,score)
            _, err := file.WriteString(term_line);
            if err != nil {
                log.Fatal(err)
            } 
        }
    }
    file.Sync()
    file.Close()
}

func main() {
    docids[0] = 1 ; docids[1] = 2; docids[2] = 3; docids[3] = 4; docids[4] = 5;
    TermMap = make(map[string] map[int] *int)
    TermInDocsMap = make(map[string] *int)
    DocWordCount = make(map[int] int)

    for i := 0; i < len(docids); i++ {
        add_terms_to_dictionary(docids[i])
    }
    sum_of_docs_per_term ();
    for i := 0; i < len(docids); i++ {
        write_tfidf_values(docids[i])
    }
}
