package adapters

import (
	"clean-arch-go/app/query"
	"clean-arch-go/common/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/task"
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirestoreTask struct {
	ID         string    `firestore:"Id"`
	Title      string    `firestore:"Title"`
	Status     string    `firestore:"Status"`
	CreatedBy  string    `firestore:"CreatedBy"`
	AssignedTo string    `firestore:"AssignedTo"`
	CreatedAt  time.Time `firestore:"CreatedAt"`
	UpdatedAt  time.Time `firestore:"UpdatedAt"`
}

type FirestoreTaskRepository struct {
	firestoreClient *firestore.Client
}

func NewFirestoreTaskRepository(firestoreClient *firestore.Client) FirestoreTaskRepository {
	if firestoreClient == nil {
		logrus.Fatalln("missing firestoreClient")
	}

	return FirestoreTaskRepository{firestoreClient: firestoreClient}
}

func (r FirestoreTaskRepository) taskCollection() *firestore.CollectionRef {
	return r.firestoreClient.Collection("tasks")
}

func (r FirestoreTaskRepository) marshalTask(tk task.Task) FirestoreTask {
	return FirestoreTask{
		ID:         tk.UUID(),
		Title:      tk.Title(),
		Status:     tk.Status().String(),
		CreatedBy:  tk.CreatedBy(),
		AssignedTo: tk.AssignedTo(),
		CreatedAt:  tk.CreatedAt(),
		UpdatedAt:  tk.UpdatedAt(),
	}
}

func (r FirestoreTaskRepository) unmarshalTask(doc *firestore.DocumentSnapshot) (task.Task, error) {
	op := errors.Op("FirestoreTaskRepository.unmarshalTask")

	model := FirestoreTask{}

	if err := doc.DataTo(&model); err != nil {
		return nil, errors.E(err, op, "unable to load document")
	}

	domainTask, err := task.From(
		model.ID, model.Title, model.Status, model.CreatedBy,
		model.AssignedTo, model.CreatedAt, model.UpdatedAt,
	)

	if err != nil {
		return nil, errors.E(op, err)
	}

	return domainTask, nil
}

func (r FirestoreTaskRepository) Add(ctx context.Context, t task.Task) error {
	collection := r.taskCollection()
	model := r.marshalTask(t)

	err := r.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Create(collection.Doc(t.UUID()), model)
	})

	if err != nil {
		return errors.E(errors.Op("FirestoreTaskRepository.Add"), err)
	}

	return nil
}

func (r FirestoreTaskRepository) UpdateByID(ctx context.Context, uuid string, updateFn func(context.Context, task.Task) (task.Task, error)) error {
	op := errors.Op("FirestoreTaskRepository.UpdateByID")

	err := r.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		documentRef := r.taskCollection().Doc(uuid)

		firestoreTask, err := tx.Get(documentRef)
		if err != nil {
			return errors.E(op, err)
		}

		domainTask, err := r.unmarshalTask(firestoreTask)
		if err != nil {
			return errors.E(op, err)
		}

		updatedTask, err := updateFn(ctx, domainTask)
		if err != nil {
			return errors.E(op, err)
		}

		if err := tx.Set(documentRef, r.marshalTask(updatedTask)); err != nil {
			return errors.E(op, err)
		}

		return nil
	})

	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (r FirestoreTaskRepository) AllTasks(ctx context.Context) ([]query.Task, error) {
	op := errors.Op("FirestoreTaskRepository.AllTasks")

	tasks, err := r.taskModelsToQuery(r.taskCollection().Query.Documents(ctx))
	if err != nil {
		return nil, errors.E(op, err)
	}

	return tasks, nil
}

func (r FirestoreTaskRepository) taskModelsToQuery(iter *firestore.DocumentIterator) ([]query.Task, error) {
	op := errors.Op("FirestoreTaskRepository.taskModelsToQuery")

	tasks := []query.Task{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		tk, err := r.unmarshalTask(doc)
		if err != nil {
			return nil, errors.E(op, err)
		}

		tasks = append(tasks, query.Task{
			UUID:       tk.UUID(),
			Title:      tk.Title(),
			Status:     tk.Status().String(),
			CreatedBy:  tk.CreatedBy(),
			AssignedTo: tk.AssignedTo(),
			CreatedAt:  tk.CreatedAt(),
			UpdatedAt:  tk.UpdatedAt(),
		})
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	return tasks, nil
}

func (r FirestoreTaskRepository) FindById(ctx context.Context, uuid string) (task.Task, error) {
	op := errors.Op("FirestoreTaskRepository.FindById")

	firestoreTask, err := r.taskCollection().Doc(uuid).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.E(op, errkind.NotExist, err)
		}

		return nil, errors.E(op, err)
	}

	domainTask, err := r.unmarshalTask(firestoreTask)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return domainTask, nil
}

func (r FirestoreTaskRepository) FindTasksForUser(ctx context.Context, userUUID string) ([]query.Task, error) {
	return []query.Task{}, errors.E(errors.Op("task.FindTasksForUser"), fmt.Errorf("dump"))
}

func (r FirestoreTaskRepository) RemoveAllTasks(ctx context.Context) error {
	op := errors.Op("FirestoreTaskRepository.RemoveAllTasks")

	collection := r.taskCollection()
	bulkWriter := r.firestoreClient.BulkWriter(ctx)

	for {
		iter := collection.Limit(400).Documents(ctx)
		numDeleted := 0

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return errors.E(op, err)
			}

			bulkWriter.Delete(doc.Ref)
			numDeleted++
		}

		if numDeleted == 0 {
			bulkWriter.End()
			break
		}

		bulkWriter.Flush()
	}

	return nil
}
