---
title: Design Twitter - System Design Interview
categories:
- English
---

[视频链接](https://www.youtube.com/watch?v=o5n85GRKuzk&t=1104s)

Let's design the **high-level architecture** of Twitter - I mean, **how hard could it be**? By the way, this video is taken from my ongoing course "System Design Interview," which will be complete by the end of this month. You can check it out on neco.io.

Before we get started, I want to mention that Twitter has been a popular topic recently, especially its **underlying infrastructure and design**. However, keep in mind that in a real interview, your design does not have to exactly match the product. It's about discussing the **trade-offs** and demonstrating your knowledge of being able to **weigh the pros and cons of an approach**. There are many similar products to Twitter, and there really isn't any one correct approach. So, we don't have to **replicate** the real Twitter design, unless, of course, you're interviewing at Twitter. In that case, you might have to,  as they recently fired everyone, so they need people to know how it works.

Let's start with the background. We know that Twitter is a social network **first and foremost**, where some people can follow other people, and that relationship can be **mutual**. But some people might end up with more followers than others. So, assume that one person is really popular, and everybody wants to read all of their tweets. But you know most people on Twitter probably aren't actually tweeting very often. Most people don't actually have many followers, including myself.

I'm mentioning this because it **hints** that this is going to be a very **read-heavy system**. Of course, on Twitter, **the whole point** is that people can create tweets. On a particular tweet, you have a person's **profile picture** and their username, and the actual content of the tweet. It can have some text, images, and videos. There are many things you can do to interact with a tweet. Of course, you can like the tweet, retweet it, **follow the person** who made the tweet, or **unfollow them**. Recently, you can even edit tweets. But this is just to give you a general idea of the **functionality** you might want to clarify with your interviewer.

Designing a **large-scale social network** like Twitter requires careful consideration of both **functional and non-functional requirements**. In a 45-minute interview, we need to focus on the most important features and aspects of the system.

**Functional requirements**:
- The first feature we want to **prioritize** is the ability for users to follow each other. This requires designing a **simple and intuitive user interface** that allows users to find and follow other users easily. 
- Another essential feature is the ability to create and post tweets. While Twitter's 140-character limit is a **well-known constraint**, we also need to consider how to handle **multimedia content** like images and videos. For example, we might allow users to **attach images or short videos to their tweets**.
- The third feature we need to design is a news feed that **displays tweets** from users that a given user is following. This feature can be more complex than it first appears, as we need to consider how to **rank tweets in the feed** and what algorithm to use. **We could use machine learning to personalize the feed for each user, but we should clarify with the interviewer whether this level of complexity is necessary for the interview**.

**Non-functional requirements**
-  **The number of users is a crucial factor to consider**. Let's assume we have 500 million total users, with around 200 million of them being  **daily active users**. This means we need to design the system to handle a massive amount of tweet reads per day, as each daily active user might read about a hundred tweets. Assuming the average tweet size is one kilobyte, this translates to around 20 billion tweet reads per day.
- **We also need to consider how much storage the system requires**. Assuming each tweet is around **one kilobyte** and that most tweets don't contain multimedia content, we can estimate that the average tweet size is around one **megabyte**. However, some tweets might contain images or videos that are much larger, so we need to plan for higher storage requirements. To estimate how much data we'll be dealing with, let's assume that there are 20 billion tweets, and each tweet takes up one megabyte of data. This means that we'll be dealing with 20 terabytes of data per day. However, since a megabyte is a million bytes, we'll need to multiply this by a thousand to get 20 petabytes per day.

In summary, designing a social network like Twitter requires careful consideration of functional and non-functional requirements. By focusing on the most critical features and aspects of the system, we can create a **scalable** and **user-friendly** platform that can accommodate millions of users and billions of tweets

Given that we're dealing with such a huge volume of data, it's clear that we'll need a storage solution that prioritizes read performance over write performance. Therefore, we can use **eventual consistency** rather than **strong consistency**.

Assuming there are 200 million daily active users, let's say that there are 50 million tweets created each day. While some users might only create one tweet per day, others might create up to ten tweets per day. Since we'll be dealing with much less data on the write side, we should focus our design on **optimizing read performance**.

**High-level architecture**
- To design our system, we'll start with the **client**, whether that's a computer or a mobile device. 
- The user will interact with the **application servers**, which will handle actions like creating a tweet, getting the news feed, and following other users. Since the news feed will be accessed frequently, we'll likely be **bottlenecked** there. We can use **stateless application servers** and **load balancers** to **scale up** our system as needed.
-  Our application servers will read from a **storage layer**, which could be a **relational database**. While a **NoSQL database** would be easier to scale, we may need to use joins to model the relationship between followers and followees, so a relational database could be the better choice. We could implement sharding to improve scalability. However, we could also store tweets and user information in a NoSQL database and use a **graph database to model the follower relationship**.
- Since we'll be dealing with massive amounts of reads, we'll need a **caching layer** to improve performance. 
- We'll also need a separate storage solution, such as **object storage**, for media files like images and videos.

When we read a tweet, we'll retrieve information such as the tweet ID, the creator of the tweet, the time it was created, and any images it included. We'll also retrieve information about the user who created the tweet, such as their profile picture.

In this Twitter design, let's assume that the HTTP request header contains an **authorization token** for us to verify the correct person who is making the tweet. However, we could also pass the user ID or the username of the person making the tweet as additional parameters in this request. To get the actual feed, we only need the user ID to retrieve the tweets of the user. However, we need to ensure that users cannot pass in other user IDs to get their feeds, which can be handled by the HTTP header authentication. **Although there is authentication going on in this system, we're not focusing on it as we're primarily looking at the Twitter design**.

In this design, we have three main interactions: making a tweet, getting the feed, and following another user. **We will be storing two things: the actual tweets and the follow relationships**. Assuming we have a relational database, we will have a table of tweets and another table of follows. The follow relationship is straightforward, with the follower being the person following another, and the followee is the person being followed.

If we want to get all the tweets of people that a particular user follows, we need to **index this table**. We would favor indexing based on the follower, which means that all records for this person will be grouped together in a particular range. If we have a read-heavy system, this approach would be ideal.

For the tweet itself, we will have the tweet ID, timestamp, user ID of the person who created it, and the content of the tweet. We will not store the media itself in the database, but we'll have a reference to that media that references the object store.

One of the bigger problems with our storage is that we will be storing a massive amount of data. If we go through the calculation mentioned earlier, roughly 50 gigabytes of data are going to be written per day to our relational database if we're not including the media. In a month, we will have 1.5 terabytes, and in a year, we'll have roughly 18 terabytes of data.

Since we will have many reads hitting this database, we can have read-only replicas of the database. If reads are the bottleneck, it's not hard to add additional database instances.

However, the problem arises when a user creates a tweet. If we have a single leader replication, all those writes are going to be hitting a single database. Ideally, we should be able to scale our writes as well, and the obvious way to do this is by using sharding. Sharding allows us to break up our database into smaller pieces to scale writes more efficiently.

To shard this Twitter design, we could do it based on user ID because that's the whole point of our design. A user only cares about a subset of users, and it doesn't make sense to do it based on tweet ID. By using the user ID as the shard key, we can break up our database into smaller pieces and ensure that a particular user only hits one of these pieces.

In case we have read-only replicas, it's okay if a single instance gets the writes and then asynchronously populates the read-only replicas with the data. If a user ends up hitting one of the replicas and gets some stale data, it's not a problem if it takes a few seconds before they get the tweet. However, if there is a sudden spike in traffic, the peak could be much higher, and it's important to be able to scale writes to handle this.


- high-level architecture
- how hard could it be？
- underlying infrastructure and design
- weigh the pros and cons of an approach
- replicate the real Twitter design
- first and foremost
- read-heavy system
- hint
	-   n. a slight or indirect indication or suggestion.
	-   v. suggest or indicate something indirectly or covertly.
-  the whole point
-  profile picture: 头像
- follow or unfollow the person
- functionality
- functional and non-functional requirements
- prioritize
- simple and intuitive user interface
- well-known constraint
- multimedia content** like images and videos
- displays tweets
- personalize the feed for each user
- daily active users
- kilobyte, megabyte, gigabyte, terabytes, petabytes
- we can create a **scalable** and **user-friendly** platform that can accommodate millions of users and billions of tweets
- eventual consistency rather than strong consistency
- stateless application servers
- load balancers
- We can use **stateless application servers** and **load balancers** to **scale up** our system as needed
- relational database,NoSQL database
- object storage
- 
- 
