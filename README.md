here comes the time for you to discover new notions and the beauty of :
•Responsive design
•DOM Manipulation
•SQL Debugging
•Cross Site Request Forgery
•Cross Origin Resource Sharing
•...
3



#### OBJECTIFS

App’s users should be able to select an image in a list of superposable images (for
instance a picture frame, or other “we don’t wanna know what you are using this for”
objects), take a picture with his/her webcam and admire the result that should be mixing
both pictures.
All captured images should be public, likeables and commentable.
4


This web project is challenging you to create a small web application allowing you to
make basic photo and video editing using your webcam and some predefined images.


Obviously, these images should have an alpha channel, otherwise your
superposition would not have the expected effect !


We will, for instance, picture the precise moment of an intergalactical cat launch,


#### HOUSE 

# -- RULEZ:

-- Fisher, probably....


1. NO ERRORS OR WARNING LOGS.    (EXCEPT:::::) any error related
to getUserMedia() are tolerated.

2. FOR SERVER use any language, but functions I use must have equvalent in PHP STANDARD LIBRARY:
- PHP STANDARD LIBRARY FUNCTIS:

3. CLIENT SIDE pages --  Client-side, your pages must use HTML, CSS and JavaScript.

no leaks etc.. 

any webserver... is ok...

 <!-- You are free to use any webserver you want, like Apache, Nginx or even the built-in
webserver1 -->


think about privacy




1. Try to structure as MVC app
2. Website is secured and form validation is implemented too


# -- APP RULEZ:

User features
• The application should allow a user to sign up by asking at least a valid email
address, an username and a password with at least a minimum level of complexity.
• At the end of the registration process, an user should confirm his account via a
unique link sent at the email address filled in the registration form.
• The user should then be able to connect to your application, using his username
and his password. He also should be able to tell the application to send a password
reinitialisation mail, if he forget his password.
• The user should be able to disconnect in one click at any time on any page.
• Once connected, an user should modify his username, mail address or password.


Gallery features
• This part is to be public and must display all the images edited by all the users,
ordered by date of creation. It should also allow (only) a connected user to like
them and/or comment them.
• When an image receives a new comment, the author of the image should be notified
by email. This preference must be set as true by default but can be deactivated in
user’s preferences.
• The list of images must be paginated, with at least 5 elements per page.



Editing features
Figure V.1: Just an idea of layout for the editing page
This part should be accessible only to users who are authenticated/connected and
should politely reject all other users who attempt to access it without being successfully
logged in.
This page should contain 2 sections:
• A main section containing the preview of the user’s webcam, the list of superposable
images and a button allowing to capture a picture.
• A side section displaying thumbnails of all previous pictures taken.
Your page layout should normally look like in Figure V.1.
• Superposable images must be selectable and the button allowing to take the picture should be inactive (not clickable) as long as no superposable image has been
selected.
• The creation of the final image (so among others the superposing of the two images)
must be done on the server side.
• Because not everyone has a webcam, you should allow the upload of a user image
instead of capturing one with the webcam.
• The user should be able to delete his edited images, but only his, not other users’
creations.


 Constraints and Mandatory things
To sum up things, your Über application should respect the following technologic choices:
• Authorized languages:
◦ [Server] Any (limited to PHP standard library)
◦ [Client] HTML - CSS - JavaScript (only with browser natives API)
• Authorized frameworks:
◦ [Server] Any (up to PHP standard library)
◦ [Client] CSS Frameworks tolerated, unless it adds forbidden JavaScript.
Your project must imperatively contain:
• One (or more) container to deploy your site with one command. anything equivalent
to docker-compose is ok.




#### BO NU SA

If the required part is done entirely and perfectly, you can add any bonus you wish; They
will be evaluated by your reviewers. You should however still respect the requirements
in the bonus parts (i.e. image processing should be done on server side).
If you lack inspiration, here are some leads:
• “AJAXify” exchanges with the server.
• Propose a live preview of the edited result, directly on the webcam preview. We
should note that this is much easier than it looks.
• Do an infinite pagination of the gallery part of the site.
• Offer the possibility to a user to share his images on social networks.
• Render an animated GIF.