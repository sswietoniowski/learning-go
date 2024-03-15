package data

import (
	"context"
	"log"

	"github.com/sswietoniowski/learning-go/projects/00_mini/04_hrms/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EmployeesMongoDBRepository struct {
	ctx        context.Context
	client     *mongo.Client
	collection *mongo.Collection
}

func NewEmployeesMongoDBRepository(ctx context.Context,
	mongoUri string, mongoDb string) (*EmployeesMongoDBRepository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("connected to MongoDB")

	const collectionName = "employees"

	repository := &EmployeesMongoDBRepository{
		ctx:        ctx,
		client:     client,
		collection: client.Database(mongoDb).Collection(collectionName),
	}

	return repository, nil
}

func (r *EmployeesMongoDBRepository) GetAll(ctx context.Context) ([]domain.Employee, error) {
	log.Println("get all employees")

	query := bson.D{}
	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, &DatabaseError{"GetAll", err}
	}

	mongoDbEmployees := make([]MongoDbEmployee, 0) // empty slice to handle empty result

	if err = cursor.All(ctx, &mongoDbEmployees); err != nil {
		return nil, &DatabaseError{"GetAll", err}
	}

	employees, err := MongoDbEmployeesToEmployees(mongoDbEmployees)
	if err != nil {
		return nil, &DatabaseError{"GetAll", err}
	}

	return employees, nil
}

func (r *EmployeesMongoDBRepository) GetById(ctx context.Context, id string) (*domain.Employee, error) {
	log.Println("get an employee by id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, &DatabaseError{"GetById", err}
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	document := r.collection.FindOne(ctx, filter)

	mongoDbEmployee := &MongoDbEmployee{}
	err = document.Decode(mongoDbEmployee)
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, &NotFoundError{ID: id}
		default:
			return nil, &DatabaseError{"GetById", err}
		}
	}

	addedEmployee, err := MongoDbEmployeeToEmployee(mongoDbEmployee)
	if err != nil {
		return nil, &DatabaseError{"GetById", err}
	}

	return addedEmployee, nil
}

func (r *EmployeesMongoDBRepository) Add(ctx context.Context, employee domain.Employee) (*domain.Employee, error) {
	log.Println("add a new employee")

	mongoDbEmployee, err := EmployeeToMongoDbEmployee(&employee)
	if err != nil {
		return nil, &DatabaseError{"Add", err}
	}

	insertResult, err := r.collection.InsertOne(ctx, mongoDbEmployee)
	if err != nil {
		return nil, &DatabaseError{"Add", err}
	}

	id := insertResult.InsertedID.(primitive.ObjectID).Hex()

	return r.GetById(ctx, id)
}

func (r *EmployeesMongoDBRepository) ModifyById(ctx context.Context, id string, employee domain.Employee) (*domain.Employee, error) {
	log.Println("modify an employee by id")

	mongoDbEmployee, err := EmployeeToMongoDbEmployee(&employee)
	if err != nil {
		return nil, &DatabaseError{"ModifyById", err}
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, &DatabaseError{"ModifyById", err}
	}

	mongoDbEmployee.ID = objectId

	filter := bson.D{{Key: "_id", Value: objectId}}
	update := bson.D{{Key: "$set", Value: mongoDbEmployee}}

	err = r.collection.FindOneAndUpdate(ctx, filter, update).Err()
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, &NotFoundError{ID: id}
		default:
			return nil, &DatabaseError{"ModifyById", err}
		}
	}

	return r.GetById(ctx, id)
}

func (r *EmployeesMongoDBRepository) RemoveById(ctx context.Context, id string) (*domain.Employee, error) {
	log.Println("remove an employee by id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, &DatabaseError{"RemoveById", err}
	}

	employee, err := r.GetById(ctx, id)
	if err != nil {
		return nil, &NotFoundError{ID: id}
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, &DatabaseError{"RemoveById", err}
	}
	if result.DeletedCount == 0 {
		return nil, &NotFoundError{ID: id}
	}

	return employee, nil
}

func (r *EmployeesMongoDBRepository) Close() error {
	if err := r.client.Disconnect(r.ctx); err != nil {
		return err
	}

	log.Println("disconnected from MongoDB")

	return nil
}
