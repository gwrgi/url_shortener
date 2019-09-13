#!/bin/bash

## 
## TEST CASE:
##
## Test that the web server is even running
##
RESULT=`curl --disable -s -X GET http://localhost:8080/v1/ping`
if [ "$RESULT" != "pong" ]; then
    echo "Actual  : '$RESULT'"
    echo "Expected: 'pong'"
    exit 1
fi
# $RESULT should be pong

urls=(
    "http://example.com" # Simple URL
    "https://example.com" # For Security
    "https://www.duckduckgo.com/" # Ends with a /
    "https://duckduckgo.com/?q=oh+the+humanity&t=ffsb&ia=web" # Includes parameters
    "https://www.google.com/search?hl=en&q=https%3A%2F%2Fduckduckgo.com%2F%3Fq%3Dlorem%2Bipsum%26t%3Dffsb%26ia%3Danswer" # Includes url encoding
    "https://longurlmaker.com/go?id=aspread%2Boutfar%2Breaching96runningv62fzShim8stretched5lengthened42TinyLinkiexpanded0Shortlinks3089eoutstretched004060Shrinkr5stretchedgohf6enduringfarawayB65012dLiteURL1196eTightURL6TraceURLloftyf04508v5UlimitxURLHawko0e6farawayrEzURL5fdShredURL0lasting2FhURL8180106URLo121lnk.in1URLvistringy071elongate3SimURLoutstretchedoutstretched65NotLongk5TraceURLlankye5outstretched8Doiopstretched1491spread%2BoutBeam.to8y10protractedeURLCuttertowering85ae08Metamarkg18p01DwarfurlenlargedmTraceURL02elongishexpanded9aEzURLTraceURL136URLPie150extensive5tallewXil6b235jrangy087lengthenedNotLongURLcutlengthy063Beam.toURLPiemTraceURLdPiURLTinyURLsr89sustained2Shim4far%2Boff140e1xpf11Shortlinkscontinued40lengthyelongated07uUrlTeaIs.gd11lengthy0stringy7x5102LiteURL0farawaybTinyLink8FwdURL11toweringFly20GetShortygreataspun%2BoutShortenURL0Sitelutionsfar%2Boffc0extensiveawdDecentURLTraceURLTightURLganglingb1Metamark52SnipURL4URLCutterShim78TinyURL0highenduringv2NanoRef6030811NotLonglastingo83413481s5MyURLstretched1c5spun%2BoutShortlinkselongated242eexpanded0aSHurla94YATUC0elongatefar%2Breachingrilasting0lingeringcontinuedShimTraceURLbl117y41g33o00Redirx92384309prolongedfar%2BreachinglastingoutstretchedFwdURLprolongedprotractedstretchinglengthened1aShredURLShim921301ShortURLm70063gangling01lengthycSitelutions9ShortURL7c9NanoRefu7Doiop09b400U769f2FwdURL166496protractedenduringm416lankyNotLongwMinilien1e0gspread%2Bout01400DigBig2elongated760115744spun%2Boutc04B65rangy3e0717o01i9lengthened14continuedTightURL23drawn%2BoutShortenURL80rangy81lengthened163elongated73StartURLstretchMetamark109PiURL1094r1Shorl2DecentURL3outstretched2rofRubyURLRedirx2spun%2Boutqrunning0d0fea4spun%2BoutShrtndfar%2Breaching11WapURLnb31Shorl45running13TightURLa2llankys0spun%2Bout6b0A2Nf8Minilien12benduring189stretchURl.ie3230continued0bfSmallrb1ShredURL301URLSitelutions70extensiveRubyURLrangyNotLong22expanded0G8L0Xilrangy00910sPiURLdistanttall01oSimURLgangling119ShrinkURLdlengthenedspun%2Bout03rangy1B65SmallrNotLong84e378gURLcut1d0c3URLvielongatek0protractedShortURL1l01Shrtnd10slongish4fShrinkURLloft" #   Really super long
)

## 
## TEST CASE:
##
## Test that each url can be made into a short url, and that the short url can be used to get the long url again
##
for url in "${urls[@]}"
do
    SHORT_URL=`curl --disable -s -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"'$url'"}'`
    LONG_URL=`curl --disable -s -w "%{redirect_url}\n" -o /dev/null -X GET $SHORT_URL`

    if [ "$LONG_URL" != "$url" ]; then
        echo "Actual  : '$LONG_URL'"
        echo "Expected: '$url'"
        exit 1
    fi
done

## 
## TEST CASE:
##
## Test number of visits
##
for url in "${urls[@]}"
do
    SHORT_URL=`curl --disable -s -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"'$url'"}'`

    # Retrieve the long url a few times
    curl --disable -s -o /dev/null -X GET $SHORT_URL
    curl --disable -s -o /dev/null -X GET $SHORT_URL
    curl --disable -s -o /dev/null -X GET $SHORT_URL
    curl --disable -s -o /dev/null -X GET $SHORT_URL

    # Now get the data
    ## TODO: Modify the URL to retrieve visit info
    ##  Call modified URL
    ##  parse response
    ##  compare visit times to the number of times the short url was expanded

done


## 
## TEST CASE:
##
## Test the time it takes both to generate the short url and the time it takes to retrieve the long url from the short url
##
for url in "${urls[@]}"
do
    SHORTEN_TIME=`curl --disable -s -w "%{time_total}" -o short_url -X POST 'http://localhost:8080/v1/create' -d '{"longUrl":"'$url'"}'`
    SHORT_URL=`cat short_url`
    EXPAND_TIME=`curl --disable -s -w "%{time_total}" -o /dev/null -X GET $SHORT_URL`

    EXPECTED_TIME=0.010

    if (( $(echo "$SHORTEN_TIME > $EXPECTED_TIME" | bc -l) )); then
        echo "ACTUAL  : $SHORTEN_TIME"
        echo "EXPECTED: $EXPECTED_TIME"
        exit 1
    fi

    if (( $(echo "$EXPAND_TIME > $EXPECTED_TIME" | bc -l) )); then
        echo "ACTUAL  : $EXPAND_TIME"
        echo "EXPECTED: $EXPECTED_TIME"
        exit 1
    fi
done

echo "Pass"
