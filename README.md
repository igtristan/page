# page
Static site generator and html tedium eliminator.


## Compilation

Download and build.  Assumes you're using brew on OSX.  Substitute for your favorite package manager or install go and git manually.

```
brew install git go 
git clone https://github.com/igtristan/page
go build .
```

At this point the "page" executable will be sitting inside the directory.
Likely you'll want to install it globally on your machine.

```
go install
```

The above command will place the executable into your $GOPATH/bin directory.
Just make sure executables can be accessed from that path on your OS.

see
https://stackoverflow.com/questions/21001387/how-do-i-set-the-gopath-environment-variable-on-ubuntu-what-file-must-i-edit/53026674#53026674


### Self Hosted Server

```
page serve -port=8080 ./directory/
```

Go to localhost:8080 in your favorite browser and view the contents of your site without precompilation.
.page files will be automatically converted to .html files


### Static Site Builder

```
page build ./src_directory ./dst_directory
```

Go through the entire src_directory recursiveley and convert all .page files to .html files and generate any tag directed artifacts

## Built in tags

### <core.html>

In order for the site generation to work correctly the root of your document must contain <core.html> instead of &lt;html&gt;.

```
<core.html>
	<head>...</head>
	<body>...</body>
</core.html>
```

### <core.include>

Include another page and tags defined within it in the current page.
Note pages that start with "_" will not be served or built.  They may only be used with the include directive.

```
<core.include page="_template.page" />

page = unix disk path to page
```

### <core.css>

Create css that is scoped to the containing element.

```
<div>
	<core.css>
		{  border: 1px solid #000;  /* only applies to the containing div  */ }
		.test { /* only applies to elements of class test within this div */ }
	</core.css>
	
	<div class="test"></div>
</div>
```


### <core.tag>

Define a user generated tag

```
<core.tag name="mytag">
	<div>
		<span>@(name)</span>
		<children />
	</div>
</core.tag>
```
usage
```
<mytag name="hi">
	<div>Some html</div>
</mytag>
```
result
```
<div>
	<span>hi/span>
	<div>Some html</div>
</div>
```

### <core.img>

Transform images on disk

```
<core.img src="myimage.png" resize="256,256,linear" />
```


## Roadmap

- Single page generation (done)
- Scoped Css (done)
- Custom Tags (done)
- Image resizing (done)
- ```<!-- comments removed from compiled html --> ``` (done)
- Favicon generation (in progress)
- Local Imports
- Remote Imports


