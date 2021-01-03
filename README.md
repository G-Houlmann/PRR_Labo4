# PRR_Labo4

# Starting the application
The binary to run the application is available in the latest release on the repo, at https://github.com/G-Houlmann/PRR_Labo4/releases/tag/v1.0 . Download the binary you need for your OS, as well as the config.json file. Specify a configuration in the config.json file and put it in the same folder as the binary. Do so on every machine specified in your configuration. The config.json file MUST be exactly the same on every machine. You can now run the binary on all the machines of your config. Don't forget to specify the id of the process as the first and only argument. Example : `./PRR_labo4 1` on linux or `./PRR_labo4.exe 1` on Windows.
The application will start being usable once every machine specified in the config has the application up and running.

Note: You can build your own binary with go build in the root directory of the repo.


## Configuration instructions  
When filling in the configuration file, please pay attention to the following: 
- The process IDs must all be lower than 1000, because of an identification specification. This specification can be changed by updating a constant in the code.
- The prime divisor attached to a process must be a prime number. All the prime number from 2 to the _Nth_ one (n being the amount of clients in the configuration file). If you skip one, the program will not work correctly.



# Test scenario
Use the `config.json` file given. In this file, process 1 has an aptitude of 5 and processes 1 and 2 have an aptitude of 15.  

- Start process 0
    - It should start an election and win it
- Immediatly Start process 1
    - It should start an election and win it
- Start process 2
    - It should start an election, but process 1 will win it because its id is lower
- Stop the process 1
    - An election should soon start on processes 0 and 2 (started by either of them), won by process 2
- Using the console, set the aptitude of process 0 to 50 (by typing `50` in the console and hitting `Enter`)
    - It should start an election and win it
- Stop all processes


### Expected results: 
Note that the amount and position of the ping and pong-related messages may differ according to your timing.  
Process 0:  
<img src="images/pr0.PNG">  

Process 1:  
<img src="images/pr1.PNG">  

Process 2:  
<img src="images/pr2.PNG">  