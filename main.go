package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

//const filePermission = 0644

type Arguments map[string]string

type ItemT struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func Perform(args Arguments, writer io.Writer) error {
	var err error = nil

	operationString := args["operation"]
	fileString := args["fileName"]

	if operationString == "" {
		return errors.New("-operation flag has to be specified")
	}
	if fileString == "" {
		return errors.New("-fileName flag has to be specified")
	}

	//perform operation
	switch operationString {
	case "list":
		err = list(fileString, writer)

	case "add":
		item := args["item"]
		if item == "" {
			err = errors.New("-item flag has to be specified")
		} else {
			err = add(item, fileString, writer)
		}

	case "remove":
		id := args["id"]
		if id == "" {
			err = errors.New("-id flag has to be specified")
		} else {
			err = remove(id, fileString, writer)
		}

	case "findById":
		id := args["id"]
		if id == "" {
			err = errors.New("-id flag has to be specified")
		} else {
			err = findById(id, fileString, writer)
		}

	default:
		err = fmt.Errorf("Operation %s not allowed!", operationString)
	}

	return err
}

func list(fileName string, writer io.Writer) error {
	bytes, err := os.ReadFile(fileName)

	if err != nil {
		err = fmt.Errorf("failed to read file: %w", err)
	} else {
		_, err = writer.Write(bytes)
		if err != nil {
			err = fmt.Errorf("failed to write bytes to file: %w", err)
		}
	}

	return err
}

func add(item, fileName string, writer io.Writer) error {
	var err error = nil
	var newUser ItemT

	err = json.Unmarshal([]byte(item), &newUser)

	if err != nil {
		return fmt.Errorf("failed to parse JSON item: %w", err)
	}

	// read users from file
	var users []ItemT
	bytes, _ := os.ReadFile(fileName)
	err = json.Unmarshal(bytes, &users)

	//if it is the first record
	if err != nil {
		//create new list and clear error
		users = make([]ItemT, 0, 1)
		err = nil
	} else {
		for _, user := range users {
			//check if added ID already exists
			//in this case we don't save anything, but just indicate the problem to console without setting error
			if user.Id == newUser.Id {
				writer.Write([]byte("Item with id " + user.Id + " already exists"))
				// here expected that "err" variable should be nil, if not - exists bug in previous code section
				return err
			}
		}
	}

	// append the new user and save the list
	users = append(users, newUser)
	err = saveFile(users, fileName)

	return err
}

func remove(id, fileName string, writer io.Writer) error {
	var users []ItemT
	//	var err error = nil

	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %s", err.Error())
	}
	err = json.Unmarshal(bytes, &users)

	if err != nil {
		return fmt.Errorf("failed to parse JSON %s", err.Error())
	}

	//create new list
	newUsers := make([]ItemT, 0, len(users))
	for _, user := range users {
		//copy to new list all users except "removed" one
		if user.Id != id {
			newUsers = append(newUsers, user)
		}
	}

	// if lenght of new slice is equal to an old one:
	//     requested ID for removal was not found
	if len(newUsers) == len(users) {
		_, err = writer.Write([]byte("Item with id " + id + " not found"))
	} else {
		err = saveFile(newUsers, fileName)
	}

	return err
}

func saveFile(users []ItemT, fileName string) error {
	bytes, err := json.Marshal(users)

	if err != nil {
		return fmt.Errorf("failed create JSON string: %w", err)
	}

	var file *os.File
	file, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		err = fmt.Errorf("failed to open file: %w", err)
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		err = fmt.Errorf("failed to write file: %w", err)
		return err
	}

	err = file.Close()
	if err != nil {
		err = fmt.Errorf("failed to close file: %w", err)
		return err
	}

	return err
}

func findById(id, fileName string, writer io.Writer) error {
	var users []ItemT

	bytes, _ := os.ReadFile(fileName)
	err := json.Unmarshal(bytes, &users)

	if err != nil {
		return fmt.Errorf("failed to parse JSON file: %w", err)
	}

	for _, user := range users {
		//user with requested ID was found
		if user.Id == id {
			bytes, err := json.Marshal(user)
			if err != nil {
				return fmt.Errorf("failed to create JSON string: %w", err)
			}

			writer.Write(bytes)
		}
	}

	return nil
}

func main() {
	// A string flag.
	operation := flag.String("operation", "defaultOperation", "")
	item := flag.String("item", "defauldItem", "")
	fileName := flag.String("fileName", "defaultFileName", "")
	flag.Parse()

	m := ItemT{}

	b := []byte(*item)
	err := json.Unmarshal(b, &m)
	if err != nil {
		fmt.Println(err)
	}
	args := Arguments{
		"id":        m.Id,
		"operation": *operation,
		"item":      *item,
		"fileName":  *fileName,
	}

	// fmt.Println(m)
	// fmt.Println(*item)

	err = Perform(args, os.Stdout)
	if err != nil {
		panic(err)
	}
}
