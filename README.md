# GET /problems/:participationID

```json
[
  {
    "id": "",
    "description": "",
    "title": "",
    "memory_limit": "(integer)KB",
    "time_limit": "(decimal)",
    "quiz_id": "",
    "number_of_test_cases": "0>(integer)<10",
  },
]
```

# GET /languages/:problemID

```json
[
  {
    "id": "integer",
    "name": "",
    "version": "decimal"
  },
]
```

# GET /examples/:problemID

```json
[
  {
    "input": "",
    "output": ""
  },
  {
    "input": "",
    "output": ""
  }
]
```
# GET /quizzes
Output:
```json
[
  {
    "ID": "90420d0a-8e34-42c5-b76c-786640858b46",
    "Title": "Rust Developer",
    "Description": "Estamos libres de PHP desde 2010!"
  },
  {
    "ID": "dff0f835-6a56-4c2d-859b-d5dfe011bfdb",
    "Title": "Fullstack Developer in Zig",
    "Description": "+10 years in experience"
  }
]
```

# GET /quizzes/:participationID
Output:
```json
[
  {
    "id": "",
    "description": "",
    "title": "",
    "memory_limit": "(integer)KB",
    "time_limit": "(decimal)",
    "quiz_id": "",
    "number_of_test_cases": "0>(integer)<10",
    "best_try": "0>(integer)<10",
    "languages": [
      {
        "id": "integer",
        "name": "",
        "version": "decimal"
      }
    ],
    "examples": [
      {
        "input": "",
        "output": ""
      },
      {
        "input": "",
        "output": ""
      }
    ]
  }
]
```

# POST /submissions/:problemID
src should be in base64 encode
Input:
```json
{
  "src" : "",
  "language_id" : ""
}
```
Output:
```json
{
  "submission_id" : ""
}
```

# PUT /submissions/:submission_id/tc/:test_case_id
Header or URL param: **X-Auth-Token=AUTH_TOKEN**

Only for judge0 results
```json
{
  "stdout": null,
  "time": "0.006",
  "memory": 2048,
  "stderr": "run: line 1:     3 Killed                  /usr/local/python-3.8.1/bin/python3 script.py\n",
  "token": "114cc4d5-4608-491f-b61e-4528ccb95480",
  "compile_output": null,
  "message": "Exited with error status 137",
  "status": {
    "id": 11,
    "description": "Runtime Error (NZEC)"
  }
}
```

# SSE /results/:submission_id
Input: 

```
event:
data: {"data": "Accepted: 3, ..."}
```

