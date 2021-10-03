## Assignment4
实现Hystrix滑动窗口熔断机制\n
1.Sever
- 模拟了Sever接受Client请求
- 模拟了Handler中调用rpc接口（qps越高，越容易失败），从而触发熔断机制，并记录到Log中
- 熔断机制参考了Hytrix三个状态的自动机，分别为OPEN、HALF_OPEN、CLOSED
- simulateResult记录了模拟结果
    - 总共触发了3次熔断
    - 熔断1s后进行HALF_OPEN试探
    - 成功后清除窗口重新运行

2.Client
- 模拟了Client发送请求
- 每1000个请求为一个周期，每100个请求增加压力