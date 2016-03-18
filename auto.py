from time import sleep
import subprocess

def main():
  files_to_commit = subprocess.check_output("git ls-files . --others --exclude-standard", shell=True).strip().split("\n")

  for f in files_to_commit:
    cmd = "git add %s; git commit -m 'update'; git push;" % f
    subprocess.call(cmd, shell=True)
    sleep(.5)

if __name__ == "__main__":
	main()
