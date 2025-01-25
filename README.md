## Domain

Build a Task Management API with role-based access, enabling two types of users: Employer and Employee.

1. Employee Role:

- View Assigned Tasks: An Employee can only view the tasks assigned to them.
- Task Status Update: An Employee can update the status of their tasks (e.g., "In Progress," "Completed").

2. Employer Role:

- Create and Assign Tasks: An Employer can create tasks and assign them to specific employees.
- View All Tasks with Filtering and Sorting:
  - Filter tasks by:
    - Assignee: View tasks assigned to a specific employee.
    - Status: View tasks based on status (e.g., "Pending," "In Progress," "Completed").
  - Sort tasks by:
    - Date: Sort tasks by creation date or due date.
    - Status: Sort tasks by task status to see active or completed tasks first.
- View Employee Task Summary: An Employer can view a list of all employees, each showing:
  - Total number of tasks assigned.
  - Number of tasks completed by each employee.
