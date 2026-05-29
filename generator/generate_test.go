package generator_test

import (
	"fmt"
	"regexp"
	"shortener/generator"
	"testing"
)

// Проверка что создается точно размером 10
func TestGenerate_Length(t *testing.T) {
	const size = 10

	for i := 0; i < 100; i++ {
		s, err := generator.Generate(size)
		if err != nil {
			t.Fatalf("Generate(%d) error: %v", size, err)
		}
		if len(s) != size {
			t.Errorf("Generate(%d) = %q, length %d, want %d", size, s, len(s), size)
		}
	}
}

// Проверка на то что создается из определенного ряда символов, указанного в тз 
func TestGenerate_Charset(t *testing.T) {
	const size = 10
	// Регулярка: ровно 10 символов из допустимого набора
	valid := regexp.MustCompile(`^[a-zA-Z0-9_]{10}$`)

	for i := 0; i < 100; i++ {
		s, err := generator.Generate(size)
		if err != nil {
			t.Fatalf("Generate(%d) error: %v", size, err)
		}
		if !valid.MatchString(s) {
			t.Errorf("Generate(%d) = %q, contains invalid characters", size, s)
		}
	}
}

// Проверка на случайности, то что разные вызовы дают различный результат
func TestGenerate_Unique(t *testing.T) {
	const size = 10
	const iterations = 1000

	seen := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		s, err := generator.Generate(size)
		if err != nil {
			t.Fatalf("Generate(%d) error: %v", size, err)
		}
		if seen[s] {
			t.Logf("collision at iteration %d: %q (rare, but possible)", i, s)
		}
		seen[s] = true
	}


	uniquePercent := float64(len(seen)) / iterations * 100
	if uniquePercent < 90 {
		t.Errorf("too many collisions: only %.1f%% unique values", uniquePercent)
	}
}

// Проверка, работает ли генератор не только с размером 10
func TestGenerate_DifferentSizes(t *testing.T) {
	sizes := []int{1, 5, 10, 20, 63}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
			s, err := generator.Generate(size)
			if err != nil {
				t.Fatalf("Generate(%d) error: %v", size, err)
			}
			if len(s) != size {
				t.Errorf("Generate(%d) = %q, length %d, want %d", size, s, len(s), size)
			}
		})
	}
}
