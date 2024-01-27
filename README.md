
GET /quizzes

GET /quizzes/:quizID

GET /quizzes/:quizID/problems/:problemID

POST /submissions/

{
  src: "",
  language_id: "",
  quiz_id: "",
  problem_id: ""
}

[
  {
    token: ""
  },
  ...
]

GET /submissions/:token

{
  "stdout": "hello, Judge0\n",
  "time": "0.001",
  "memory": 376,
  "stderr": null,
  "token": "8531f293-1585-4d36-a34c-73726792e6c9",
  "message": null,
  "status": {
    "id": 3,
    "description": "Accepted"
  }
}