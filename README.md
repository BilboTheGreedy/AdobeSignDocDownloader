# AdobeSignDocDownloader
Utility to download and sync documents to filesystem from Adobe Sign. Built to run on Windows. Paths/Filenames are set to be Windows friendly.

## Build cache data
Build local cache store in order to download.

```bash
.\AdobeSignDocDownloader.exe -cache
```
## Download documents 
Download documents. Set a hefty max concurrency with max flag like 50.

```bash
.\AdobeSignDocDownloader.exe -max 50
```



## Usage/Examples
print output to console and debug flags exist.
Console flag removes progressbar. 
By default all "SIGNED" documents are downloaded. If you want any other, specify status flag.

```bash
.\AdobeSignDocDownloader.exe -max 100 -console -debug -status "CANCELLED"
```