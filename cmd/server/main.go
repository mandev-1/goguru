package main

import (
	"fmt"
	"camagru/internal/database"
	"camagru/internal/router"
	"log"
	"net/http"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	defer db.Close()

	router.SetupRoutes(db)

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

<td width='30'></td>
</tr>
</tbody>
</table>
</div>
</body>
</html>`
}
