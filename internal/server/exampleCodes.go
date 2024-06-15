package server

import "fmt"

func ExampleCode(languageID int32) (string, error) {
	exampleCode, ok := codeExamples[languageID]
	if !ok {
		return "", fmt.Errorf("language %d not supported", languageID)
	}
	return exampleCode, nil
}

var codeExamples = map[int32]string{
	60: `// Código de ejemplo en Go:
// Recibir una palabra (stdin) e imprimir la palabra invertida (stdout).

package main

import "fmt"

func main() {
   var input string
   fmt.Scanln(&input)

   runes := []rune(input)
   for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
       runes[i], runes[j] = runes[j], runes[i]
   }

   output := string(runes)
   fmt.Print(output)
}`,

 54: `// Este es un código de ejemplo en C++ que resuelve el siguiente problema:
// Recibir una palabra (stdin) e imprimir la palabra invertida (stdout).

#include <iostream>
#include <string>
#include <algorithm>

int main() {
   std::string input;
   std::getline(std::cin, input);

   std::string output = input;
   std::reverse(output.begin(), output.end());

   std::cout << output;
   return 0;
}`,

	62: `// Este es un código de ejemplo en Java que resuelve el siguiente problema:
// Recibir una palabra (stdin) e imprimir la palabra invertida (stdout).
import java.util.Scanner;

public class Main {
   public static void main(String[] args) {
       Scanner scanner = new Scanner(System.in);
       String input = scanner.nextLine();

       StringBuilder sb = new StringBuilder(input);
       String output = sb.reverse().toString();

       System.out.print(output);
       scanner.close();
   }
}`,

	71: `# Este es un código de ejemplo en Python que resuelve el siguiente problema:
# Recibir una palabra a través de la entrada estándar (stdin) e imprimir la palabra invertida en la salida estándar (stdout), sin mostrar ningún otro mensaje adicional.

input_value = input()
output_value = input_value[::-1]
print(output_value, end='')`,

	63: `// Este es un código de ejemplo en JavaScript que resuelve el siguiente problema:
// Recibir una palabra (stdin) e imprimir la palabra invertida (stdout).
const readline = require('readline').createInterface({
 input: process.stdin,
 output: process.stdout
});

readline.question('', (input) => {
 const output = input.split('').reverse().join('');
 console.log(output);
 readline.close();
});`,
}
