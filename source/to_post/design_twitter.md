


let's design the high level architecture of Twitter I mean how hard Could It Be 

by the way this video is taken from my ongoing course system design interview which will be complete by the end of this month you can check it out on

neco.io before we get started 

I do want to mention Twitter has been quite a popular topic recently especially the underlying infrastructure and design but keep in mind that in a real interview Your Design does not have to exactly match the product that's not what it's about at all 

it's about discussing the trade-offs and kind of demonstrating your knowledge of being able to weigh the pros and cons of an approach and of course there's many similar products to Twitter 

there really isn't any one correct approach so we don't have to actually replicate the real Twitter design unless of course you'reinterviewing at Twitter in that case you might have to because they recently fired everyone so they need people to know how it works 

so let's start with the background we know that Twitter is a social network **first and foremost** where some people can follow other people and that relationship can be be mutual. this erson can also follow the other person

but some people might end up with more followers than others right 

so assume that one person is really popular and everybody wants to read all of their tweets but you know most people on Twitter probably aren't actually

tweeting very often most people don't actually have many followers including myself.  

I'm mentioning this because it kind of hints that this is going to be a very **read heavy system** and of course on Twitter the whole point is that people can create tweets so on a particular tweet you have a person like this is their **profile picture** and their username and then the actual content of the tweet. it can have some text it can have some images and it can have a video there's a lot of things you can do to interact with a tweet of course you can like the Tweet,you can do a retweet, you can follow the person who actually made the Tweet or unfollow them recently you can even edit tweets now but this is just to give you a general idea of what kind of **functionality** you might want to clarify with your **interviewer** 

so now actually digging into the functional requirements

Twitter is very very large of course we can't design every little piece of functionality in a 45 minute interview

so what exactly do we want to spend most of our time on and what parts can we just kind of hand wave and dismiss and just briefly discuss 
let's say the first feature the priority feature is that we want to be able to follow other users. so users can follow each other now there's no point of following other people if you can't actually create tweets so that's also going to be just as important and then third is actually viewing a news feed now at a basic level. these two features are pretty simple but viewing a feed can definitely be more complicated especially when we get into how we want to rank that feed what kind of algorithm are we going to use.probably there's machine learning going on in there and in many cases you end up seeing Tweets in your feed by people that you're not even following but we're gonna assume that that's not the case for reviewing a feed we just want to see
tweets of people that we actually follow.I think that would be something worth clarifying with your interviewer that's something that can kind of **scope down** this interview because you might assume that we're doing something really complicated but your interviewer is actually looking for something more simple that's a trap that you don't want to fall into now what exactly is going to go in the Tweet itself

we know that Twitter actually has a limit on tweet size I think it's 140 characters 
let's assume that your interviewer gives you that number but at the same time with social networks of course we end up with images and videos and let's say that yes we are going to include these in our design 

so now let's transition into the **non-functional requirements** which is not going to be completely separate from these actual features 
we're implementing the first thing you probably want to know is how many users we're talking about here
let's say the number is 500 million total users but in terms of about here let's say the number is 50 daily active users 
we have about 200 million of them that are daily active 

I think that's pretty close to the real number but remember the real number isn't so important I think the main observation here is that almost half half of the users are active so when we actually create feeds for users we'll be doing it for most users most people are going to be logging on most people are going to be viewing their feed well not most but nearly half which is a pretty large percentage about 40 percent Now

While most people will be viewing their feed they probably won't be creating tweets but again this is something we have to clarify let's say of those 200
million daily active users each of them will read about a hundred tweets per day. so 200 million times a hundred that's going to be 20 billion tweet reads per day now what is the size of each tweet if we have 140 characters that's about 140 bytes but let's assume that there's additional information with a tweet we 140 bytes but let's assume that there have the username of that tweet and possibly there's a lot more metadata to be safe or just a basic tweet that just includes text we can assume that for each tweet we have to do a **kilobyte** of reading from our storage now we alsoknow that some tweets can contain image 140 bytes but let's assume that there' s and videos so on average this is going to be higher how much higher we could spend a lot of time digging into the math you could ask your interviewer a bunch of questions but most likely this is not what they want you to spend time on let's just average this up to a **megabyte** because maybe videos on average are 10 megabytes if they're a bit longer which I don't know what the limit is for a video length on Twitter but it could be reasonably high but we also know that few tweets are going to actually have this so we average it down to this because most tweets are going to be about a kilobyte let's say so let's say a megabyte per each tweet 

so how much data are we going to be reading if it's 20 billion tweets a megabyte for each tweet that is quite a lot of data.so if each tweet was just one byte 20 billion that's going to be 20 **Gigabytes**. but now we're actually multiplying this by a **megabyte** which is a million so multiply this by a thousand we get 20 **terabytes**. but we have to multiply it by a thousand again because that's you know what a million is and then we get to 20 **petabytes** so overall we're going to be reading 20 petabytes of data per day now it's no surprise to us that this is going to be a read heavy system. 

this is kind of hinting to us what type of storage solution should we use we probably don't need to be **strongly consistent**  **eventual consistency** is enough and that brings us to how much are we going to be writing per day how many tweets are we going to be creating per day. well we have 200 million daily active users so let's say a reasonable number is 50 million tweets created per day most people aren't going to be creating tweets but maybe some people create 10 tweets per day so this is a **decent number** but you're not going to be guessing this your interviewer should be giving you something reasonable now we could go through the rest of the math with this number and I can show you that we're going to be writing much less than 20 petabytes of data per day especially if we don't include the images and videos which we're probably not going to be directly storing in a database but the main thing here to realize is that yes **we're going to be writing much less than we're reading** so that's how we want to optimize our design and let's say that the average user follow those about a hundred people so a hundred follows per person but of course there can be power users who have a thousand or ten thousand people that they follow but the more important consideration here is for a user how many followers can they have someone like Kim Kardashian 

I don't know if she's the most popular on Twitter but I think it's at least over a hundred million followers that they have so this is the more important consideration people who have a massive amount of followers so the question is going to be for all the people that follow Kim Kardashian how are they gonna get the tweets this is kind of hinting to us that wherever we're storing her tweets it's gonna get overloaded pretty quickly so now that we kind of know what we're looking for. 


     let's get into the high level design we know of course that everything is going to start with our client whether that's a computer or a mobile device it doesn't really matter for us. we're focusing on the back end which is agnostic to the front end. we know the first thing our user is going to be hitting is the application servers to perform actions like creating a tweet or getting their news feed or following someone now because of the scale that we're dealing with. we're probably going to be **bottlenecked** by getting the news feed that's what's going to be happening most frequently and if we want to scale this up assuming that these application servers are **stateless** it should be easy to scale them up and of course we will have the **load balancer** in between this that's pretty hand wavy I mean that's something you can just memorize and say well if you want to scale horizontally. scale this and put a load balancer in there it's pretty trivial so I'm not going to spend a lot of time on that 

      Now of course our application server is going to be reading from some storage let's say we do have a database and let's say that it is a **relational database** and you might be thinking if we're going to be doing read heavy why use a relational database why not just have a **nosql database** well it depends on what type of data we're going to be

storing do we need joins in this case

and we could because we do have a very

relational model when it comes to

following that's a pretty clear

relationship between followers and

follow ease so that's a reason to go

with a relational database now in theory

you it would be easier to scale a nosql

database but we can Implement sharding

with a relational database so that does

give us some flexibility though after

finishing our high level design we might

want to revise this because we could

just store tweets and user information

in a nosql database and then have the

graph DB which would be very easy to

find that follower relationship because

a graph DB is essentially like an

adjacency list graph where every person

is like a node in a graph and to find

all the people that they follow you just

have to look at every outgoing Edge and

to find all the followers of a person

you just have to look at every incoming

Edge so depending on your expertise and

your background and of course what your

interviewer is looking for and what they

might be familiar with you can kind of

have some discussion about these

differences now with the massive amount

of reads that we're going to be doing we

basically have to have a caching layer

in between so as we're reading tweets we

will be hitting our cache before we hit

our database but also remember that we

are going to be storing media so we need

a separate storage solution for that

media related National databases aren't

the best for that so we'll have some

type of object storage for that

something like Google Cloud Storage or

Amazon S3 so when we actually read a

tweet we'll be getting the information

about that tweet like the Tweet ID who's

the creator of that tweet what time was

it created whether it included an image

or not what was the image that it

included the profile picture of the

person who made it the application

server can then fetch the image like the

profile picture or the video that showed

up in that tweet and it can do that

separately but at the same time because

these assets are static in nature it may

be better to actually distribute them

over a CDN Network so then actually our

application server does not have to

interact with the object storage the

application server will respond to the

user with all the information that they

need including the URL of that image or

video that they need and then the client

whether they're using a mobile device or

a desktop browser will make a separate

request which will actually hit our CDN

Network which is tied to our object

storage what type of algorithm would we

use in this case well even though we're

looking at the high level right now we

probably want to use a poll based CDN we

don't want to necessarily push every

image or video to the CDN immediately

also remember the benefit of a CDN is

that it's geographically located close

to the user we know that people in India

might be looking at different types of

tweets and images and videos than people

in the United States so it doesn't make

sense to put every single new tweet push

it directly to the CDN Network and with

a pull-based CDN we kind of guarantee

that the stuff that's loaded onto our

CDN is the relevant stuff that people

want to see anyway the popular things so

now let's spend most of our time

actually digging into the details which

some people like to start out with the

interface that we'll be using we will

have a couple so remember we have a

create tweet there could be a lot of

metadata sent with that request of

course the user ID of the person

creating the Tweet but mainly the user

is actually responsible for sending the

content of the tweets so one is the

actual text and then second is going to

be the actual media of course every

tweet has a created timestamp but we

assume that that will be handled server

side and every tweet has to be

identified but we'll assume that the

Tweet ID is also created server side and

of course the user ID of the person

that's actually creating it well we'll

assume that in the header of the HTTP

request that there's some authorization

token for us to know that the correct

person is making the Tweet but we could

also have the user ID passed into this

request or the username of the person

and next of course we have getting the

actual feed and that really doesn't need

any information at all that's a very

basic read request we don't need to send

additional data that we're going to be

actually storing get feed should just

pass in the user ID so that we know

which person's feed are we getting but

at the same time I should not be able to

pass in your user ID even if I know it

to get your user feed there's nothing

that can go super wrong with that but it

shouldn't be allowed and that would be

handled by the HTTP header we know that

that there's actual authentication going

on in this system it's just that we're

not focusing on that we're focusing on

the actual Twitter design authentication

happens with pretty much every

application and we also have the follow

interaction so a user can follow another

person they'll pass in their user ID and

maybe the username of the person that

they're trying to follow so these are

the three main interactions now how are

we actually going to be storing this

data in particular we're going to be

storing two things the actual tweets so

assuming we have a relational database

we'll have a table of tweets and we'll

also have a table of follows so the

follow relationship is pretty simple you

have the follow e the person who's

following the other person that's going

to be a string and we'll have another

the actual follow ER I think I misspoke

when I was describing followey so the

follower is the person that's following

the following the follow e is the one

that's being followed now assuming that

this is a table and remember for a user

we want to get all the tweets of people

that this person follows so they're the

follower and they want all the tweets of

their followers how would we Index this

table we'd probably want to index based

on the follower because then all records

in that table will be grouped together

based on the follower so all the people

that this guy follows will be grouped

together in the table it'll just be a

range query so assuming this is our

table all records for this person will

be grouped together in a particular

range over here I think it's worth

mentioning that we would favor indexing

based on the follower if we have a read

heavy system now for the Tweet itself we

kind of briefly talked about it of

course we're gonna have the Tweet ID

we're going to have the time stamp we're

gonna have the user ID of the person who

created it and we're gonna have the

content of the Tweet whether it's the

text which technically could be empty if

we have some media attached to it but we

actually won't be storing the media

itself in the database we'll have a

reference to that media that references

the object store now the bigger problem

with our storage is we're going to be

storing a massive amount of data if you

went through the calculation that I

mentioned earlier you would have gotten

to I think roughly 50 gigabytes of data

are going to be written per day to our

relational database if we're not

including the media so that's a lot of

data in a month we'll have I think 1.5

terabytes in the course of a year will

have roughly 18 terabytes so it is a

large amount of data but it's actually

reasonable and the good thing about

Twitter is that tweets are relatively

small but the problem is we're going to

be having so many reads hitting this

database a person is going to be reading

a hundred tweets per day and we have 200

million of them so the first approach

and the obvious thing to do is to have

read only replicas of this database if

reads are the bottleneck it's not hard

to just add additional database

instances now the problem will be when a

user actually creates a tweet if we have

single leader replication all those

rights are going to be hitting a single

database and remember if we have 50

million writes per day you divide that

by a hundred thousand which is roughly

the amount of seconds in a day we get

500 rights per second well probably even

more than that but at the same time

there's gonna be moments where we have

traffic this is the average but Peak

could be much higher I mean Peak could

be even 10 times this amount depending

on what's going on maybe Elon Musk does

something crazy so ideally we should be

able to scale our rights as well and the

obvious way to do this is by using

sharding but the question is how are we

going to be implementing this sharding

before we get into that I forgot to

mention if we do have read-only replicas

it's okay if a single instance gets the

rights and then asynchronously populates

the read-only replicas with the data

because in the case that a user ends up

hitting one of the replicas and gets

some stale data it's okay if it takes

five seconds after a tweet is created

before a user actually gets that tweet

or it might even take 20 seconds that

would wouldn't be ideal but it's not the

end of the world with something like

Twitter now the question is how are we

actually going to be sharding this what

type of Shard key are we going to be

using I think the most obvious and easy

way would be to do it based on user ID

because that's kind of the whole point

of our design the way we scoped it out a

user only cares about a subset of users

they don't care about every user it

doesn't make a lot of sense to do it

based on tweet ID because what we want

is to break up our database into pieces

and we want a particular user to ideally

only have to hit one of these pieces

with the logic of our system especially

the sharding logic a user should know

which people they follow and then we can

route the request to the appropriate

shards that contain the tweets of the

people that they follow and we'll know

that because we can use our Shard key to

determine that but if we break up the

tweets based on tweet ID we actually

don't know which Shard contains the

tweets of the people that they follow so

we'd have to query all the shards that

kind of defeats the purpose so we'll be

choosing to do this based on user ID and

since we don't actually have complex

queries and joins that we're doing to

actually retrieve the tweets how will we

get the people that they follow well if

we Shard that based on user ID as well

all the people that this guy follows

should be on A Single Shard now a

potential problem is that we will

actually have to order the tweets just

fetching the tweets is not enough we

will actually have to query them and

then order them based on the time that

they were created and of course how many

tweets are we looking for if we're just

looking for a small number like 20 what

if the user wants to scroll down in

their feed and we want 20 more tweets we

don't necessarily need to get all tweets

immediately we need to wait for the user

to actually scroll or maybe we can

actually get all the tweets immediately

but before I even get into that let's

quickly look at how our current system

is going to work when a user creates an

essentially it will go through the

server based on the user ID we will find

the appropriate Shard and then store the

tweet on that Shard and then any images

and media on object storage as users

view their feed we may have to query

multiple shards to find all the relevant

tweets and then order them and then send

them back to the user that could

definitely be very slow that's clearly

the bottleneck here now we do have a

caching layer and in theory the most

popular tweets will be stored already on

this caching layer and if they're not

then we can have some type of lru

algorithm working here because we care

about the most recent tweets most likely

people aren't going to be viewing tweets

from a year ago even if they had like a

billion views after a while people get

tired of them lru might be better here

and in theory caching should definitely

help us lower the latency but remember

different people will have different

tweets we definitely can't guarantee

that all 20 tweets that this person

wants to see are already going to be

cached what if 19 of them are already

cashed and then that last one tweet we

still have to go and read disk to get

that tweet and while you're scrolling

maybe all your tweets are loaded but one

tweet in the middle is taking a few

extra seconds that's not a great user

experience so the problem we're running

into is not really scalability it's

latency and caching helps with that and

sharding also helps with that that's

kind of the point of sharding with

read-only replicas we're able to handle

scale but as we break up the data into

smaller chunks then we can also lower

latency because we're going to be

querying a smaller amount of data but we

still may have to query multiple shards

to get this latency even lower we can

get pretty creative and we can actually

generate the news feed of users

asynchronously and we would do that for

every single user in theory because we

know a large amount of the users 200

million out of 500 million are actually

active it makes sense to generate the

news feed for all of them even if 60

percent of them aren't going to actually

view it it's not a ton of wasted work

and we could also tune it such that we

only regenerate the news feed for people

that are actually active within the last

30 days now at a very high level what we

can do is have some kind of message

queue or Pub sub system which will take

every new tweet that is created will

also not just be written to the database

but it will be sent from the app servers

to the pub sub queue and this queue will

feed into a cluster something like a

spark cluster but the point is that

these workers will in parallel process

all the messages that we're getting

which will include every time there's a

new tweet that is generated there could

be a lot so we need to do this

asynchronously and the point of this is

that these will basically feed into a

new cache and this cache will be

responsible for actually storing the

user feed so now when a user loads their

home page and gets their list of 20

tweets the application server will will

actually be hitting this feed cache

maybe this cache is used for individual

tweets that are maybe embedded on

websites or sometimes you just open up

an individual tweet not in the context

of a user feed and you want to maybe

look at the replies of that tweet but

this is starting to make a little bit

less sense or even having this cache so

I'm just kind of giving some possible

use cases but this is the cache that

will actually have a feed for every

single user so if this is in memory it

may have a large amount of data now we

have 200 million users and if we want to

have 100 tweets for every single user it

could be a very large amount so we

probably want to Shard this similar to

how we did with our relational database

but the point is that this will

definitely lower the latency because the

feeds are not generated as a user

actually requests to actually get the

feed we don't have to actually run a

complex query on our relational database

having to query multiple shards and then

joining the results together based on

the time that they were created and then

ordering them like that the feed will

actually all already be created now the

complicated part is actually updating

the Tweet we kind of went through the

flow of when a tweet is created it will

be added to the message queue and then

for that individual tweet these workers

will add that tweet to all the feeds of

people that are following the author of

that tweet but now the problem is if

somebody has a hundred followers then

these workers will have to update a

hundred feeds but what about somebody

like Kim Kardashian who has maybe a

hundred million followers updating a

hundred million feeds every single time

somebody popular like that makes a tweet

is very very expensive maybe in that

case it's not the end of the world for

us to actually have to update the feed

of that user at the time that they

actually request it big somebody could

have a hundred million followers and us

having to update 100 million feeds every

time they make a tweet is pretty

expensive but not all 100 million of

those followers are even loading their

feed every single day so it'd probably

be easier to do that work as it's needed

so when a user makes a request they get

their feed but maybe in parallel to that

our app server could look for other

tweets probably a popular tweet by

somebody like that is already cached

here so then at that point our

application server would also update the

feed of that user now what gets even

more complicated is what happens when a

user follows somebody new their feed has

to be updated in that case as well now

it's okay we can tolerate a few seconds

maybe five or ten seconds before their

feed is actually updated but it does

have to get updated nonetheless and we

would have sort of a similar mechanism

where a user follows somebody new we add

a message to a message queue we have a

bunch of workers that then have to

update the feed cache of the person who

just followed somebody else and by the

way this cluster of course is going to

actually have to read our database as

well which actually has the tweets

themselves remember the whole point of

doing this is to lower the latent agency

when somebody opens up their home page

and they want to see 20 tweets we want

it to be as quick as possible and all

this complexity arises from that if we

can tolerate a few more seconds we can

simplify our design but as long as this

discussion has already been I want to

say that all of this has actually been

pretty high level things can get much

more complicated and this is actually

definitely not exactly how Twitter is

designed there are actually problems

with this design as well there could be

concurrent updates to our cache there's

a lot of details that we haven't

discussed how would the ordering of the

feed even be in this case when somebody

loads 20 tweets they want to see 20 more

we obviously have to paginate that and

there's a ton more details that we could

have gone into the main takeaway though

here is that Twitter definitely cannot

be designed or built in a weekend

designing these large-scale systems is a

very complicated task even Engineers who

worked at Twitter for years still ran

into issues where they had to modify

designs and if you read some of the

white papers from 2 2010 up until now

they actually did they changed parts of

their database how they were storing

relations how they were storing who

follows who so with that I encourage you

to read some of the official papers that

the engineering teams at Twitter have

written if you'd like to learn more

because while things can get complicated

they are still pretty interesting

