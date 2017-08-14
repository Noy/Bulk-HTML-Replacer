# Go-HTML-Replacer
Asynchronously go through multiple directories and changing HTML Tags, Css Classes and many more!

##### Instructions to use

#### Clone the Repo:

```
git clone https://github.com/Noy/Go-HTML-Replacer
```

#### This is the default changeFile func

```go
func changeFile(file io.Reader) (newContent string, err error) {
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return
	}
	doc.Find("title").Each(func(i int, sel *goquery.Selection) {
		text := sel.Text()
		sel.ReplaceWithHtml("<title>" + strings.ToLower(text) + "</title>")
	})
	newContent, err = doc.Html()
	return
}
```
You can change the selection/goquery to anything you want

#### Make sure you set the correct directory by using the `````--source <directory>````` flag

The output of the result is very user-friendly. You'll know exactly what's going on :D

e.g.

```
    [+] reading all files in dir "." to modify
    [+] ready with 1 items 
    [2.99ms] processed "test\test.html"
    [0.00s] processed all items, wrote 1 updates
```

#### Interesting point:
	- When I ran the task over more than 13,000 directories, which within them, 
	  contained more, this was our result from the print statement above:
	- [161.04s] processed all items, wrote 27046 updates
	- That roughly translates to 2 minutes and 41 seconds. 
	  Changing 27,046 .html files.
	- Asynchronously completing this task (parallel) took just under 3 minutes.
	  Imagine how long it'd take recursively?

##### Using the [GoQuery]("https://github.com/PuerkitoBio/goquery") lib

##### Credit to [Joey]("https://github.com/Twister915") for the help. (he's the best)

### Hope this fites your needs!
