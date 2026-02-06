### These are instructions on how to extract the images to add to Diablo 2 items ###

- Every file in this folder contains a type of item:
    - base: contains html with all base items
    - misc: contains html with the images of runes, gems, potions, throwing potions, quest items, consumables, crafted items and misc.
    - set: contains html with all set items
    - uniques: contains html with all unique items

- How to extract?
We will use html selectors, this is an example of how an item is being displayed:

<article class="element-item element-item-vt z-dbvfpad wideling z-sold-transition" loading="lazy"><span class="z-hidden">Hammers</span><a href="/uniques/stone-crusher-t937.html"><div class="z-graphic z-graphic-2x3 lozad" data-background-image="/styles/zulu/theme/images/items/stonecrusher_graphic.png" style="background-repeat:no-repeat;background-size:56px 84px;width:56px;height:84px;display:inline-block;"></div></a><h3 class="z-sort-name"><a href="/uniques/stone-crusher-t937.html" class="z-uniques-title">Stone Crusher</a></h3><h4>Elite Unique<br><div class="zi zi-smalltopic zi-smalltopic-nbsicon zi-baseitem lozad" data-background-image="/styles/zulu/theme/images/items/warhammer_ticon.png" style="background-repeat:no-repeat;"></div><span class="ajax_catch"><a href="/base/legendary-mallet-t1728.html" data-href="/ajax.php?var=1728" class="z-white ajax_link">Legendary Mallet</a></span></h4><p class="z-smallstats"><span class="zso_defense z-hidden">0</span><span class="zso_throwdamage z-hidden">0</span><span class="zso_twohdamage z-hidden">0</span><span class="z-white">1H damage:</span> <span class="zso_onehdamage">190-256</span><br><span class="z-white">Base speed:</span> <span class="zso_basespeed">20</span><br><span class="z-white">Durability:</span> <span class="zso_durability">65</span><br><span class="z-white">Req Strength:</span> <span class="zso_rqstr">189</span><br><span class="zso_rqdex z-hidden">0</span><span class="z-white">Req level:</span> <span class="zso_rqlevel">68</span><br><span class="z-white">Quality level:</span> <span class="zso_qualitylvl">76</span><br><span class="z-white">Treasure class:</span> <span class="zso_trclass">84</span><br><span class="zso_maxsock z-hidden">0</span><span class="zso_baseblock z-hidden">0</span>+<code title="Variable stat or range" class="z-trusty z-trusty-code z-trusty-code-var">280-320</code>% Enhanced Damage<br>Damage +<code title="Variable stat or range" class="z-trusty z-trusty-code z-trusty-code-var">10-30</code><br>+50% Damage To Undead<br>-25% Target Defense<br>40% Chance of Crushing Blow<br>-100 To Monster Defense Per Hit<br>+<code title="Variable stat or range" class="z-trusty z-trusty-code z-trusty-code-var">20-30</code> To Strength</p><div class="z-vf-hide"><span class="z-smallstats"><a href="/viewforum.php?f=2/#filter=Patch%201.10" class="z-uniques-title">Patch 1.10 or later</a></span><br></div><span class="z-vf-br"></span><div class="z-forum-deets z-hidden"><div class="z-grey">Views: <span class="z-sort-views">28023</span><span class="z-space"></span><span class="z-space"></span>Likes: <span class="z-sort-likes">3</span><span class="z-space"></span>Comments: <span class="z-sort-comments">3</span></div><a href="https://diablo2.io/post4099890.html#p4099890" title="Go to last post" class="z-grey">Last post: <a href="https://diablo2.io/member/Bizington/" style="color:#cfd0fe;" class="username-coloured">Bizington</a></a><span class="z-grey">, <time datetime="2026-01-27T13:03:52+00:00"><span class="z-relative-date" title="Tue Jan 27, 2026 1:03 pm">2 days ago</span></time><span class="z-sort-lastpost z-hidden">1769519032</span></span></div></article>

Every item is contained inside and article with class starting with "element-item".
The first anchor tag (<a href="/uniques/stone-crusher-t937.html"><div class="z-graphic z-graphic-2x3 lozad" data-background-image="/styles/zulu/theme/images/items/stonecrusher_graphic.png" />) contains the image we need to store. The base url is https://diablo2.io/. So you can access every image with https://diablo2.io//styles/zulu/theme/images/items/stonecrusher_graphic.png.


You should create a small program, inside catalog-api, in diablo 2 context folders to:

- Read every html and look for the images we need to display the items
- Download the image and store to Supabase Storage (we are running supabase locally)
- Update all items imageUrl property the link of the correspondent image
- Some items may contain the same image
- Save all the entities their images in database
- Update all catalog queries to return the item image property
