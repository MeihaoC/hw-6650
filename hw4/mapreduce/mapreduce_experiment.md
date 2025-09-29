# MapReduce Word Count Experiment

## What I Built

I created a system that counts words in Shakespeare's Hamlet by distributing the work across multiple ECS tasks. Instead of sequential processing, three workers count different sections simultaneously, then a reducer combines their results.

The system has 5 parts:
- **Splitter**: Split the text into 3 equal chunks
- **Mapper 1, 2, 3**: Three workers that each count words in their chunk
- **Reducer**: Combines all the counts into a final answer

Everything runs on AWS using Docker containers and stores files in S3.

## Performance Analysis

### Speed Improvement:
- **Sequential Pipline**: 0.88 seconds
- **Parallel Pipeline**: 0.62 seconds  
- **Speed boost**: 1.42x faster (42% improvement)
- **Time saved**: 0.26 seconds

The system counted 29,604 total words and found 4,803 different words in the text.

## Why Only 1.42x Faster with 3 Workers?

I expected 3x faster with 3 workers, but got 1.42x because:
- The splitter and reducer still run one at a time (can't parallelize everything)
- Time spent uploading/downloading files from S3
- Network delays between services
- The text file is small (116KB), so overhead costs matter more

## What Was Hard?

### Managing IP Addresses:
- Each service got a random IP address from AWS
- Had to manually track 5 different IPs
- IPs changed every time I restarted a service
- Lots of copy-paste errors

### No Automatic Recovery:
- If one mapper died, everything failed
- Had to manually restart and retry
- No way to know if a service was healthy

### Manual Coordination:
- Ran each step by hand with curl commands
- Had to wait and watch for each step to finish
- Copy URLs between commands (prone to typos)

## What I Learned

1. **Parallel processing works!** Even with overhead, running tasks simultaneously saves time.

2. **Distributed systems are complex.** Coordinating multiple services manually is error-prone and tedious.

3. **Small files don't benefit as much.** The overhead of splitting and combining takes significant time relative to processing 116KB.

4. **Infrastructure matters.** Without service discovery, load balancing, and orchestration, managing distributed systems is painful.

## If I Did It Again

- Use a task queue (SQS) instead of direct HTTP calls
- Add automatic retries for failed mappers
- Use service discovery instead of hardcoded IPs
- Test with larger files (GB+) where parallelization shines
- Add health checks and auto-restart for failed tasks

## Conclusion

My MapReduce system successfully demonstrates distributed computing - it splits work, processes in parallel, and combines results. The 42% speedup proves the concept works, even though manual coordination was challenging. This experiment helped me understand why tools like Kubernetes and Apache Spark exist - they solve the coordination problems I faced.