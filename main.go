package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// const filePermission = 0644

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
		err = errors.New("-operation flag has to be specified")
	} else if fileString == "" {
		err = errors.New("-fileName flag has to be specified")
	} else {
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
	// unmarshal the new user
	var newUser ItemT
	err = json.Unmarshal([]byte(item), &newUser)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal the new user JSON: %w", err)
	} else {
		// read users from file
		var users []ItemT
		bytes, _ := os.ReadFile(fileName)
		err = json.Unmarshal(bytes, &users)

		//if it is the first record
		if err != nil {
			users = make([]ItemT, 0, 1)
		} else {
			for _, user := range users {
				//check if added ID already exists
				if user.Id == newUser.Id {
					writer.Write([]byte("Item with id " + user.Id + " already exists"))
					return nil
				}
			}
		}
		// append the new user and save the list
		users = append(users, newUser)
		err = saveUsers(users, fileName)
		if err != nil {
			err = fmt.Errorf("failed to save users: %w", err)
		}
	}
	return err
}

func remove(id, fileName string, writer io.Writer) error {
	var users []ItemT
	var err error = nil
	bytes, _ := os.ReadFile(fileName)
	err = json.Unmarshal(bytes, &users)
	if err != nil {
		// if file is empty do nothing
		err = nil
	} else {
		newUsers := make([]ItemT, 0, len(users)-1)
		for _, user := range users {
			if user.Id != id {
				newUsers = append(newUsers, user)
			}
		}

		if len(newUsers) == len(users) {
			writer.Write([]byte("Item with id " + id + " not found"))
			err = nil
		} else {
			err = saveUsers(newUsers, fileName)
			if err != nil {
				err = fmt.Errorf("failed to save users: %w", err)
			}
		}
	}

	return err
}

func saveUsers(users []ItemT, fileName string) error {
	bytes, err := json.Marshal(users)
	if err != nil {
		err = fmt.Errorf("failed to marshal users as JSON: %w", err)
	} else {
		var file *os.File
		file, err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			err = fmt.Errorf("failed to open file for adding: %w", err)
		} else {
			_, err = file.Write(bytes)
			if err != nil {
				err = fmt.Errorf("failed to write data to file: %w", err)
			} else {
				err = file.Close()
				if err != nil {
					err = fmt.Errorf("failed to close file after adding: %w", err)
				}
			}
		}
	}
	return err
}

func findById(id, fileName string, writer io.Writer) error {
	var users []ItemT
	bytes, _ := os.ReadFile(fileName)
	err := json.Unmarshal(bytes, &users)
	if err == nil {
		for _, user := range users {
			if user.Id == id {
				bytes, err := json.Marshal(user)
				if err != nil {
					return fmt.Errorf("failed to marshal user as JSON: %w", err)
				}

				writer.Write(bytes)
			}
		}
	}

	return err
}

func main() {
	// A string flag.
	operation := flag.String("operation", "defaultOperation", "startup message")
	item := flag.String("item", "defauldItem", "startup message")
	fileName := flag.String("fileName", "defaultFileName", "startup message")
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
