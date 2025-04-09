package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("nothing was found")
)

type Employee struct {
	Name     string
	Position string
}
type Sth struct {
	Name string
}

type Manager struct {
	Employee
	Sth
	Workers []Employee
}

func main() {
	manager := Manager{}
	fmt.Println(manager.Employee.Name)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c := make(chan int)
	go func(ctx context.Context) {
		time.Sleep(2 * time.Second)
		c <- 123
	}(ctx)

	if x := 123 * 321; x > 100 {

	}

	select {
	case x := <-c:
		fmt.Println("x = ", x)
		break
	case <-ctx.Done():
		fmt.Println("oh no", ctx.Err())
	}

	// wg := &sync.WaitGroup{}
	// wg.Add(2)
	// go func() {
	// 	defer wg.Done()
	// 	c <- 3
	// }()

	// go func() {
	// 	defer wg.Done()
	// 	x := <-c
	// 	fmt.Println("x =", x)
	// }()

	// wg.Wait()
}

func mess() {
	fmt.Println("afdsa")
	err := ErrNotFound
	err = fmt.Errorf("database call: %w", err)

	if errors.Is(err, ErrNotFound) {
		fmt.Println("not found")
	}
	original := errors.Unwrap(err)
	fmt.Println(original)

	fmtrs := []Fmt{&HtmlFormatter{}, MarkDownFormatter{}}
	for _, f := range fmtrs {
		fmt.Println(f.Header(3, "Hello worlrd"))
		fmt.Println(f.ListItem(1, "first"))
		fmt.Println(f.ListItem(2, "second"))
		fmt.Println(f.ListItem(3, "third"))
		fmt.Println(f.Paragraph("asdfafda"))
	}
}

type MarkDownFormatter struct {
}

// Header implements Fmt.
func (m MarkDownFormatter) Header(level int, text string) string {
	return fmt.Sprintf("%s %s", strings.Repeat("#", level), text)
}

// ListItem implements Fmt.
func (m MarkDownFormatter) ListItem(index int, text string) string {
	return fmt.Sprintf("%d. %s", index, text)
}

// Paragraph implements Fmt.
func (m MarkDownFormatter) Paragraph(text string) string {
	return fmt.Sprintf("\n%s\n", text)
}

type Fmt interface {
	ListItem(index int, text string) string
	Header(level int, text string) string
	Paragraph(text string) string
}

type HtmlFormatter struct{}

// Header implements Fmt.
func (hf *HtmlFormatter) ListItem(index int, text string) string {
	return fmt.Sprintf("<li>[%d] %s</li>", index, text)
}

func (hf *HtmlFormatter) Header(level int, text string) string {
	return fmt.Sprintf("<h%d>%s</h%d>", level, text, level)
}
func (hf *HtmlFormatter) Paragraph(text string) string {
	return fmt.Sprintf("<p>%s</p>", text)
}

type task struct {
	a int
	b int
}

type result struct {
	task
	out int
}

func producer(tasks chan<- task) {
	const N = 1000
	for i := 0; i < N; i++ {
		tasks <- task{a: i * 10000, b: i*20000 + (i%100)*1000}
	}

	close(tasks)
}

func consumer(tasks <-chan task, results chan<- result) {
	for task := range tasks {
		var res int
		for i := task.a; i < task.b; i++ {
			res += i
		}
		results <- result{task: task, out: res}
	}
}

func concurrencyStuff() {
	tasks := make(chan task)
	results := make(chan result)
	go producer(tasks)

	wg := sync.WaitGroup{}
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			consumer(tasks, results)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	i := 1
	for res := range results {
		fmt.Printf("%d %#v\n", i, res)
		i += 1
	}
}

func sliceStuff() {

	nums := [3]int{1, 2, 3}
	fmt.Printf("%#v\n", nums)

	fmt.Println(len(nums))
	fmt.Println(cap(nums))

	// xs := make([]int, 0)
	xs := []int{1, 2, 3, 4, 5, 6, 7}

	fmt.Printf("%#v\n", xs)
	fmt.Printf("len = %d cap= %d\n", len(xs), cap(xs))

	xs = append(xs, 8)

	fmt.Printf("%#v\n", xs)
	fmt.Printf("len = %d cap= %d\n", len(xs), cap(xs))

	xs = append(xs, 1, 2, 3, 4, 5, 6, 7)

	fmt.Printf("%#v\n", xs)
	fmt.Printf("len = %d cap= %d\n", len(xs), cap(xs))

	ys := xs[2:5]
	fmt.Printf("%#v\n", ys)
	fmt.Printf("len = %d cap= %d\n", len(ys), cap(ys))
	ys = append(ys, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1)

	fmt.Printf("%#v\n", xs)
	fmt.Printf("len = %d cap= %d\n", len(xs), cap(xs))

	fmt.Printf("%#v\n", ys)
	fmt.Printf("len = %d cap= %d\n", len(ys), cap(ys))

	zs := make([]int, len(xs))
	copy(zs, xs)
	zs = append(zs, ys...)

	fmt.Printf("%#v\n", zs)
	fmt.Printf("len = %d cap= %d\n", len(zs), cap(zs))

	m := make(map[string]int)
	m["bob"] += 1
	fmt.Printf("%#v", m)

}
