# gsfa
Program to retrieve AtCoder sample cases

## Install  
`$ go get -u github.com/tkhsy2/gsfa`

This application uses [chromedriver](https://sites.google.com/a/chromium.org/chromedriver/).  
Install chromedriver if it is not already installed.

## Usage

`$ gsfa [contest name]`  


## Sample  

`$ gsfa abcXXX`

The above command creates a "./gsfa/abcXXX" directory and a sample I/O case file.  
In most cases, XXX is actually a number.

**output image**

    gsfa                     
    └─ abcXXX                
        ├─ A                 
        │   ├─ A_2_in.txt    
        │   └─ A_2_out.txt   
        │   
        ├─ B                 
        │   ├─ B_1_in.txt    
        │   └─ B_1_out.txt   
        │   
        ├─ C                 
        │   ├─ C_1_in.txt    
        │   └─ C_1_out.txt   
        │   
        └─ D                 
            ├─ D_1_in.txt    
            ├─ D_1_out.txt   
            ├─ D_2_in.txt    
            └─ D_2_out.txt   
            
